package main

import (
	"errors"
	"fmt"
	"sort"
)

type TreeNode struct {
	leafnode bool

	parent *TreeNode
	left   *TreeNode //valid if and only if leafnode is false
	right  *TreeNode //valid if and only if leafnode is false

	byteid byte  //valid if and only if leafnode is true
	freq   int32 //if not leaf node, sum of left.freq, right.freq
}

/*
	String format: "(leaf/parent	[byteid	frequency: leftchild rightchild]\n"
	NOTE: If a node is not nil, then its string representation will end with a newline.
	If a node is nil, then its string representation will just be <nil>, or whatever the
	Go default happens to be, if it's ever changed.
	This was a stylistic choice.
*/
func (n *TreeNode) String() string {
	if n.leafnode {
		return fmt.Sprintf("(leaf   [%3d %7d: %v %v])\n", n.byteid, n.freq, n.left, n.right)
	} else {
		return fmt.Sprintf("(parent [%3d %7d: ...  ...])", n.byteid, n.freq) +
			fmt.Sprintf("  \t%v", n.left) +
			fmt.Sprintf("  \t%v", n.right)
	}
}

type Tree struct {
	root *TreeNode
}

/*
	Given a slice of Node Pointers, this function will return a slice one size smaller
	so long as no index rules are violated.

	Index rules: index1 must be distinct from index2, index1 and index2 must be in the
	bounds of nodeslice

	The node slice will not contain the pointers at indices index1 and index2, but will
	contain a new node pointer whose children are the nodes pointed to by the pointers
	at index1 and index2. The returned slice will be sorted by the frequency of the
	referenced node as indicated by its freq field.

	TODO: If time: convert the forest from a slice to a linked list
*/
func Combine(nodeslice []*TreeNode, index1 int, index2 int) ([]*TreeNode, error) {
	//holds size of parametrized nodeslice. This is used just so that we don't call len() all the time.
	var origslice_size int
	//holds size of newly allocated retslice. This is used just so that we don't call len() all the time.
	var retslice_size int
	//used for iterating through origslice when we copy values from origslice to retslice
	var origslice_index int
	//used for iterating through retslice when we copy values from origslice to retslice
	var retslice_index int
	//smaller slice we will return
	var retslice []*TreeNode
	//these are pointers for the nodes at index1 and index2, and the node into which they will be combined
	var n1, n2, nreplace *TreeNode

	//initialize size variables
	origslice_size = len(nodeslice)
	retslice_size = origslice_size - 1
	//the latter two checks work because we use >, not >=
	if index1 == index2 || index1 > retslice_size || index2 > retslice_size {
		return nil, errors.New(ERRMSG_ENCODING_TREE_DUPLICATE_INDICES)
	}

	n1 = nodeslice[index1]
	n2 = nodeslice[index2]

	//this is the ACTUAL combination
	nreplace = new(TreeNode)
	//set the attributes of combination node
	nreplace.leafnode = false
	nreplace.freq = n1.freq + n2.freq
	n1.parent = nreplace
	n2.parent = nreplace

	//as per the spec, the smallest node must be the new node's left child
	if NodeLess(n1, n2) {
		nreplace.left, nreplace.right = n1, n2
	} else {
		nreplace.left, nreplace.right = n2, n1
	}

	//allocate new node 1 size smaller
	retslice = make([]*TreeNode, retslice_size)
	//copy elements in
	for retslice_index = 0; retslice_index < retslice_size && origslice_index < origslice_size; origslice_index++ {
		if nodeslice[origslice_index] == n1 || nodeslice[origslice_index] == n2 {
			continue
		}
		retslice[retslice_index] = nodeslice[origslice_index]
		retslice_index++
		// new_slice_index++
	}
	retslice[retslice_size-1] = nreplace

	/*Sort them.

	TODO: if time: rework the above loop so that we don't have to call sort.Slice() here.*/
	sort.SliceStable(retslice, func(i int, j int) bool {
		return retslice[i].freq <= retslice[j].freq
	})
	return retslice, nil
}

//if `a` and `b` are determined to be equivalent then this function will return true.
func NodeLess(a *TreeNode, b *TreeNode) bool {
	//equal frequency case
	//lowest byteid is "less". Trivial decision. Consistency matters.
	return a.freq < b.freq
}
