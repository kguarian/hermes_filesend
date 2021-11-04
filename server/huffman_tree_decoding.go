package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
)

func ReadFrequencyTable(f *os.File) (encoding_decision bool, frequency_table []int32, err error) {
	var curr_byte byte
	var curr_byte_offset int8
	var inverted_field int32
	var true_field int32
	var field_sz byte
	var field_index int16
	var field_bit_offset byte
	var reader *bufio.Reader

	frequency_table = make([]int32, 256)
	reader = bufio.NewReader(f)

	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}

	curr_byte, err = reader.ReadByte()
	if err != nil {
		return
	}

	field_sz = curr_byte >> 3

	if field_sz == 0 {
		return
	} else {
		encoding_decision = true
	}

	fmt.Printf("FieldSz = %d\n", field_sz)

	curr_byte <<= 5
	curr_byte >>= 5

	curr_byte_offset = 5

	for field_index = 0; field_index < 256; field_index++ {
		true_field = 0
		for field_bit_offset = 0; field_bit_offset < field_sz; field_bit_offset++ {
			inverted_field <<= 1
			if ((curr_byte >> (8 - curr_byte_offset - 1)) % 2) == 1 {
				inverted_field++
			}
			curr_byte_offset++
			if curr_byte_offset == 8 {
				curr_byte, err = reader.ReadByte()
				if err != nil {
					return encoding_decision, nil, err
				}
				curr_byte_offset = 0
			}
		}
		for field_bit_offset = 0; field_bit_offset < field_sz; field_bit_offset++ {
			true_field <<= 1
			if inverted_field%2 == 1 {
				true_field++
			}
			inverted_field >>= 1
		}
		frequency_table[field_index] = true_field
	}
	if err != nil {
		return
	}

	Info_Log(frequency_table)
	return
}

func DecodeFile(ref_file *os.File, tgt_file *os.File, bit_offset byte, ref_offset int, huffmantree *Tree, freq_slice []int32, encoding_decision bool) (err error) {
	var w *bufio.Writer
	var r *bufio.Reader
	var root *TreeNode
	var traversal_node *TreeNode
	var total_byte_count int64
	var byte_index int64
	var read_byte byte
	var writequeue chan byte
	var ghost_freqslice []int32

	//for debugging
	var currtraversal_length int
	var intermediate_traversal_encoding uint64

	var bytes_retrieved int

	writequeue = make(chan byte, QUEUESIZE)
	ghost_freqslice = make([]int32, 256)
	_, err = tgt_file.Seek(0, 0)
	Errhandle_Exit(err, ERRMSG_SEEK)
	w = bufio.NewWriter(tgt_file)
	r = bufio.NewReader(ref_file)

	if encoding_decision {
		offset_printval, err := ref_file.Seek(int64(ref_offset), 0)
		Errhandle_Exit(err, ERRMSG_SEEK)
		fmt.Printf("offset: %d, byte pos: %d\n", offset_printval, bit_offset)

		root = huffmantree.root
		traversal_node = root

		for i := 0; i < 256; i++ {
			total_byte_count += int64(freq_slice[i])
		}
		read_byte, err = r.ReadByte()
		if err != nil {
			Errhandle_Exit(err, ERRMSG_READ)
		}
		for byte_index = 0; byte_index < total_byte_count; {
			intermediate_traversal_encoding <<= 1
			intermediate_traversal_encoding += uint64(read_byte >> (8 - bit_offset - 1) % 2)
			currtraversal_length++
			if (read_byte>>(8-bit_offset-1))%2 == 1 {
				if traversal_node.right != nil {
					traversal_node = traversal_node.right
				} else {
					return errors.New(ERRMSG_DECODING_NIL_NODE)
				}
			} else {
				if traversal_node.left != nil {
					traversal_node = traversal_node.left
				} else {
					return errors.New(ERRMSG_DECODING_NIL_NODE)
				}
			}

			if traversal_node.leafnode {
				if *logflag {
					var debugslice []byte
					Info_Log(fmt.Sprintf("%x ", traversal_node.byteid))
					if strconv.IsPrint(rune(traversal_node.byteid)) {
						Info_Log(fmt.Sprintf("%c\t\t", traversal_node.byteid))
					}
					debugslice = make([]byte, currtraversal_length)
					for i := 0; i < currtraversal_length; i++ {
						if intermediate_traversal_encoding>>(int64(currtraversal_length-i-1))%2 == 1 {
							debugslice[i] = '1'
						} else {
							debugslice[i] = '0'
						}
					}
					fmt.Printf("char: %2x, length: %2d, traversal: %20b\n", traversal_node.byteid, currtraversal_length, ((intermediate_traversal_encoding)<<(64-currtraversal_length-1))>>(64-currtraversal_length-1))
					Info_Log(debugslice)
				}
				writequeue <- traversal_node.byteid
				if len(writequeue) == QUEUESIZE {
					for len(writequeue) != 0 {
						w.WriteByte(<-writequeue)
					}
				}
				ghost_freqslice[traversal_node.byteid]++
				if ghost_freqslice[traversal_node.byteid] > freq_slice[traversal_node.byteid] {
					line, err := ref_file.Seek(0, 1)
					Errhandle_Exit(err, ERRMSG_SEEK)
					if *logflag {
						fmt.Printf("Found too many of byte %d at byte %d of the reference file. This will be byte index %d @ Traversal length: %d, offset %d. Traversal: %b\n", traversal_node.byteid, line, byte_index, currtraversal_length, bit_offset, intermediate_traversal_encoding)
					}
				}
				// fmt.Printf("%d\n", traversal_node.byteid)
				currtraversal_length = 0
				traversal_node = root
				byte_index++
			}
			bit_offset++
			if bit_offset == 8 {
				read_byte, err = r.ReadByte()
				if err != nil {
					if byte_index == total_byte_count {
						err = nil
						break
					}
					log.Printf("failed after %d bytes of decoding, %d bytes read\n", byte_index, bytes_retrieved)
					Errhandle_Exit(err, ERRMSG_READ)
				}
				bytes_retrieved++
				bit_offset = 0
			}
		}
		// println(len(writequeue))
		for len(writequeue) != 0 {
			w.WriteByte(<-writequeue)
		}
		w.Flush()
		if *logflag {
			Info_Log(fmt.Sprintf("byte index: %d\tbytes retrieved: %d\n", byte_index, bytes_retrieved))
		}
		Info_Log(ghost_freqslice)
		if !reflect.DeepEqual(freq_slice, ghost_freqslice) {
			for i, _ := range freq_slice {
				if ghost_freqslice[i] != freq_slice[i] {
					fmt.Printf("Mismatch on character %d=%x=%c: actual: %d contrived: %d\n", i, i, i, freq_slice[i], ghost_freqslice[i])
				}
			}
			log.Println(freq_slice)
			err = errors.New(ERRMSG_DECODING)
			return err
		}
	} else {
		var bytecount int64
		_, err = ref_file.Seek(1, 0)
		Errhandle_Exit(err, ERRMSG_SEEK)
		for read_byte, err = r.ReadByte(); err == nil; read_byte, err = r.ReadByte() {
			w.WriteByte(read_byte)
		}
		w.Flush()
		if err == io.EOF {
			err = nil
		}
		println(bytecount)
	}

	return
}
