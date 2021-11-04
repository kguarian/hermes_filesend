package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const QUEUESIZE int = 8092

func CreateFrequencyTable(ref *os.File) (freq_slice []int32, err error) {
	var curr_byte byte
	var reader *bufio.Reader

	//this is our frequency table
	freq_slice = make([]int32, 256)
	_, err = ref.Seek(0, 0)
	if err != nil {
		Errhandle_Exit(err, ERRMSG_SEEK)
	}

	reader = bufio.NewReader(ref)
	//read each byte, mark its frequency in the table
	for curr_byte, err = reader.ReadByte(); err == nil; curr_byte, err = reader.ReadByte() {
		freq_slice[curr_byte]++
	}
	//end-of-field is expected; nothing else is.
	if err != io.EOF {
		return nil, err
	} else {
		err = nil
	}
	return
}

/*
	CalculateHeaderSize predicts the size of the header, assuming that the program decides to encode the file with a full header.

	The header is structured as follows:
	1) 5 bits to indicate the length of each field (the frequency of a particular byte) -- Big Endian, packed left
	2) 256*value of the first 5 bits -- Little Endian, packed left

	--END OF HEADER--

	Endianness was chosen as follows (this should help you build a decoding method):
	1) The length of the largest byte frequency is chosen as the length of EVERY field.
	2) It is written in a single byte, then shifted left by 3 places. This is Big-Endian by nature.
	3) The rest of the header is written. Since we have the frequencies, which can be as long as 32 bytes a piece, we
	just shift left and and the bit, repeatedly, until the byte is full. Then we enqueue the byte for writing, and this
	produces a little-endian pattern.

	TODO: refactor to move from calculating length of each frequency to calculating length of only one frequency
*/
func CalculateHeaderSize(freqtable []int32) (byte, int) {
	//first return value
	var maxfreqlength int32
	//second return value, ultimate goal of this function
	var headerlength int
	//used to track current frequency in table.
	var currfreq int32
	//used to allow us to calculate the length of the frequency
	var currfreqoffset int32

	//find greatest length of frequency
	maxfreqlength = 0

	for _, v := range freqtable {
		currfreqoffset = 31
		currfreq = v
		for currfreqoffset = 31; currfreqoffset > 0; currfreqoffset-- {
			currfreq >>= currfreqoffset
			if currfreq%2 == 1 {
				break
			} else {
				currfreq = v
			}
		}
		if currfreqoffset+1 > int32(maxfreqlength) {
			maxfreqlength = currfreqoffset + 1
		}
		//Info_Log(v)
	}
	headerlength = int(256*maxfreqlength + 5)
	return byte(maxfreqlength), headerlength
}

/*
	CompressionDecider compares the projected size of the file if we copied and added to it a minimal header against the projected size
	of the file if we compressed it by the Huffman Encoding Algorithm.
*/
func CompressionDecider(freqtable []int32, nodeslice []*TreeNode, tree *Tree, fieldsz byte, origfile *os.File) (bool, int64, []byte, error) {
	var encodinglengthtable []byte = make([]byte, 256)
	var copiedfilesz int64
	//projectedfilesz measures BIT count
	var encodedfilesz int64
	var seeklocation_beforecall int64
	var filereader *bufio.Reader
	var currbyte byte
	var nodedepth byte
	var treeroot *TreeNode
	var currnode *TreeNode
	var retsz int64
	var err error

	var charcount int

	seeklocation_beforecall, err = origfile.Seek(0, 1)
	Errhandle_Exit(err, ERRMSG_SEEK)

	//calculate size of copy encryption
	copiedfilesz, err = origfile.Seek(-1, 2)
	Errhandle_Exit(err, ERRMSG_SEEK)

	//+2 for header, +1 for 0-indexing to 1-counting
	copiedfilesz += 3

	//length of each header entry (header for header) will be written.
	encodedfilesz = (5 + 256*int64(fieldsz))
	_, err = origfile.Seek(0, 0)
	Errhandle_Exit(err, ERRMSG_SEEK)
	filereader = bufio.NewReader(origfile)

	//here, we go through the motions of compression by traversing the tree as we would during compression
	//and counting the number of bits our encoding requires. We convert to bytes below.
	treeroot = tree.root

	for currbyte, err = filereader.ReadByte(); err == nil; currbyte, err = filereader.ReadByte() {
		if freqtable[currbyte] != 0 {
			if encodinglengthtable[currbyte] == 0 {
				currnode = nodeslice[currbyte]
				nodedepth = 0

				for ; currnode != treeroot; currnode = currnode.parent {
					nodedepth++
				}
				encodinglengthtable[currbyte] = nodedepth
			}
		} else {
			err = errors.New(ERRMSG_ENCODING_FREQTABLE_INVALID_ZERO)
			Errhandle_Exit(err, ERRMSG_ENCODING_HEADER_DECISION)
			return false, 0, encodinglengthtable, err
		}
		encodedfilesz += int64(encodinglengthtable[currbyte])
		charcount++
	}

	if err != io.EOF {
		Errhandle_Exit(err, ERRMSG_READ)
	}

	//bits to bytes
	//remainders take up an extra byte
	if encodedfilesz%8 != 0 {
		encodedfilesz += 8
	}
	encodedfilesz /= 8

	if err != io.EOF {
		Errhandle_Exit(err, ERRMSG_READ)
	}

	//TODO: Strongly consider silencing this log message before submission
	log.Printf("field sz: %d, size if encoded: %d, size if copied: %d, encode decision: %s\n", fieldsz, encodedfilesz, copiedfilesz, func() string {
		if encodedfilesz < copiedfilesz {
			return "COMPRESS"
		} else {
			return "COPY"
		}
	}())

	//Seek back to original read location (on function call)
	_, err = origfile.Seek(seeklocation_beforecall, 0)
	Errhandle_Log(err, ERRMSG_SEEK)
	if encodedfilesz < copiedfilesz {
		retsz = encodedfilesz
	} else {
		retsz = copiedfilesz
	}
	return encodedfilesz < copiedfilesz, retsz, encodinglengthtable, nil
}

/*
	Here, we just write the header. See details in CalculateHeaderSize's function header.
*/
func WriteHeader(freq_slice []int32, fieldsz byte, encodedecision bool) (data []byte, currbyte byte, bit_offset byte, err error) {

	if fieldsz == 0 || !encodedecision {
		data = make([]byte, 1)
		return
	}

	var data_offset int = 0

	var field_size_precise float32 = float32(256) * float32(fieldsz) / float32(8)
	if field_size_precise-float32(int(field_size_precise)) != 0 {
		data = make([]byte, int(field_size_precise+1))
	} else {
		data = make([]byte, int(field_size_precise))
	}

	currbyte = fieldsz
	bit_offset = 5

	Info_Log("field size: " + strconv.Itoa(int(fieldsz)))
	Info_Log(encodedecision)

	// SetConsoleColor(RED)
	// log.Printf("%d\n", currbyte)
	// SetConsoleColor(RESET)

	for _, v := range freq_slice {
		// if fieldsz < header_major_sz {
		for i := byte(0); i < fieldsz; i++ {
			//add bit
			currbyte <<= 1
			if (v>>i)%2 != 0 {
				currbyte += 1
			}
			bit_offset++
			//write byte if it's full
			if bit_offset == 8 {
				data[data_offset] = currbyte
				data_offset++
			}
		}
	}
	Info_Log(bit_offset)
	return

}

/*
	EncodeFile retraces the steps of CompressionDecider, but actually writes the bytes.
	EncodeFile writes in the little endian, packed left, format described in CalculateHeaderSize's function header.
*/
func EncodeBytes(encoding_tree *Tree, ns []*TreeNode, encodinglengthtable, data, write_in_progress []byte, tgtfile *io.Writer, writeheader bool, writequeue_ptr *chan byte, currbyte, offset byte) (retData []byte, err error) {
	var currnode *TreeNode
	var lastnode *TreeNode
	var currreadbyte byte
	var currentIndex int
	var f float32
	var currtraversal_length uint8
	var writequeue chan byte = *writequeue_ptr
	var charcount int
	var encoding int64
	var w *bufio.Writer

	w = bufio.NewWriter(*tgtfile)
	currtraversal_length = 0

	Info_Log(writequeue)
	Info_Log(len(writequeue))
	Errhandle_Exit(err, ERRMSG_SEEK)

	if !writeheader {
		//return:
		return bytes.Join([][]byte{make([]byte, 1), data}, nil), nil
	} else {
		f = (float32(write_in_progress[0]>>3)*256 + 5)
		currentIndex = int(f / 8)
		currbyte = data[currentIndex]
		for _, currbyte = range data {
			// fmt.Printf("NEW LOOP: Character: %d\n", currreadbyte)
			//Info_Log(ns)

			currnode = ns[currbyte]
			for lastnode, currnode, currtraversal_length = currnode, currnode.parent, 0; lastnode != currnode; lastnode, currnode, currtraversal_length = currnode, currnode.parent, currtraversal_length+1 {
				if lastnode == currnode.right {
					//slotting in 1's for the huffman encoding
					encoding |= 1 << currtraversal_length
				}
			}
			//fmt.Printf("inverted encoding: %b, ln: %d\n", inverted_encoding, currtraversal_length)
			if *logflag {
				var debugslice []byte
				debugslice = make([]byte, currtraversal_length)
				for i := byte(0); i < currtraversal_length; i++ {
					if encoding>>i%2 == 1 {
						debugslice[i] = '1'
					} else {
						debugslice[i] = '0'
					}
				}
				Info_Log(string(debugslice))
			}
			for i := byte(0); i < currtraversal_length; i++ {
				currbyte <<= 1
				if encoding%2 == 1 {
					currbyte++
				}
				encoding >>= 1
				offset++
				if offset == 8 {
					writequeue <- currbyte
					if len(writequeue) == QUEUESIZE {
						for len(writequeue) != 0 {
							w.WriteByte(<-writequeue)
						}
					}
					offset = 0
				}
			}

			if currtraversal_length != encodinglengthtable[currreadbyte] {
				log.Printf("traversal length: %d, correct length: %d, byte id: %d\n", currtraversal_length, encodinglengthtable[currreadbyte], currreadbyte)
				return nil, errors.New(ERRMSG_ENCODING)
			}
			charcount++
		}
		for len(writequeue) != 0 {
			w.WriteByte(<-writequeue)
		}
		Info_Log(offset)
		if offset != 0 {
			w.WriteByte(currbyte << (8 - offset))
		}
		if *logflag {
			Info_Log(fmt.Sprintf("offset: %d\n", offset))
		}
		err = w.Flush()
		Errhandle_Exit(err, ERRMSG_FLUSH)
		Info_Log(charcount)

		return
	}
}
