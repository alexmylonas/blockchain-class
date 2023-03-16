package merkle

// Copyright 2017 Cameron Bergoon
// Licensed under the MIT License, see LICENCE file for details.

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Hashable represents the behavior concrete must exhibit to be used in the tree.
type Hashable[T any] interface {
	Hash() ([]byte, error)
	Equals(other T) bool
}

type Tree[T Hashable[T]] struct {
	Root         *Node[T]
	Leafs        []*Node[T]
	hashStrategy func() hash.Hash
	MerkleRoot   []byte
}

func WithHashStrategy[T Hashable[T]](hashStrategy func() hash.Hash) func(*Tree[T]) {
	return func(t *Tree[T]) {
		t.hashStrategy = hashStrategy
	}
}

func NewTree[T Hashable[T]](values []T, options ...func(t *Tree[T])) (*Tree[T], error) {
	var defaultHashStrategy = sha256.New
	t := Tree[T]{
		hashStrategy: defaultHashStrategy,
	}
	for _, option := range options {
		option(&t)
	}

	if err := t.Generate(values); err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *Tree[T]) Generate(values []T) error {
	if len(values) == 0 {
		return errors.New("no values provided")
	}

	var leafs []*Node[T]
	for _, value := range values {
		hash, err := value.Hash()
		if err != nil {
			return err
		}
		leafs = append(leafs, &Node[T]{
			Hash:  hash,
			Value: value,
			leaf:  true,
			Tree:  t,
		})
	}

	if (len(leafs) % 2) != 0 {
		duplicate := &Node[T]{
			Hash:  leafs[len(leafs)-1].Hash,
			Value: leafs[len(leafs)-1].Value,
			leaf:  true,
			dup:   true,
			Tree:  t,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return err
	}
	t.Root = root
	t.Leafs = leafs
	t.MerkleRoot = root.Hash

	return nil
}

// Rebuild is a helper function that will rebuild the tree, only using the
// data that it currently holds in leaves.
func (t *Tree[T]) Rebuild() error {
	var data []T
	for _, leaf := range t.Leafs {
		data = append(data, leaf.Value)
	}
	if err := t.Generate(data); err != nil {
		return err
	}
	return nil

}

// PROOOF
func (t *Tree[T]) Proof(data T) ([][]byte, []int64, error) {
	for _, node := range t.Leafs {
		if !node.Value.Equals(data) {
			continue
		}

		var merkleProof [][]byte
		var order []int64
		nodeParent := node.Parent

		for nodeParent != nil {
			if bytes.Equal(nodeParent.Left.Hash, node.Hash) {
				merkleProof = append(merkleProof, nodeParent.Right.Hash)
				order = append(order, 1) // right leaf, concat second
			} else {
				merkleProof = append(merkleProof, nodeParent.Left.Hash)
				order = append(order, 0) // left leaf, concat first
			}
			node = nodeParent
			nodeParent = nodeParent.Parent
		}

		return merkleProof, order, nil
	}

	return nil, nil, fmt.Errorf("data not found in tree")
}

// Verify verifies the proof of a given data
func (t *Tree[T]) Verify() error {
	calculatedMerkleRoot, err := t.Root.verify()
	if err != nil {
		return err
	}

	if !bytes.Equal(calculatedMerkleRoot, t.MerkleRoot) {
		return fmt.Errorf("calculated merkle root does not match the original one")
	}

	return nil
}

// VerifyData
func (t *Tree[T]) VerifyData(data T) error {
	for _, node := range t.Leafs {
		if !node.Value.Equals(data) {
			continue
		}

		currentParent := node.Parent
		for currentParent != nil {
			rightBytes, err := currentParent.Right.CalculateHash()
			if err != nil {
				return err
			}
			leftBytes, err := currentParent.Left.CalculateHash()
			if err != nil {
				return err
			}

			h := t.hashStrategy()
			if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
				return err
			}
			if !bytes.Equal(h.Sum(nil), currentParent.Hash) {
				return fmt.Errorf("invalid merkle proof")
			}
			currentParent = currentParent.Parent
		}

		return nil
	}

	return fmt.Errorf("data not found in tree")
}

func (t *Tree[T]) Values() []T {
	var values []T
	for _, leaf := range t.Leafs {
		if leaf.dup {
			continue
		}
		values = append(values, leaf.Value)
	}
	return values
}

// RootHex converts the merkle root byte hash to a hex encoded string.
func (t *Tree[T]) RootHex() string {
	return hexutil.Encode(t.MerkleRoot)
}

func (t *Tree[T]) String() string {
	s := ""
	for _, leaf := range t.Leafs {
		s += fmt.Sprint(leaf)
		s += "\n"
	}

	return s
}

func (t *Tree[T]) MarshalText() ([]byte, error) {
	panic("do not marhsal the tree, use values instead")
}

type Node[T Hashable[T]] struct {
	Hash   []byte
	Value  T
	Tree   *Tree[T]
	leaf   bool
	dup    bool
	Left   *Node[T]
	Right  *Node[T]
	Parent *Node[T]
}

func (n *Node[T]) verify() ([]byte, error) {
	if n.leaf {
		return n.Hash, nil
	}

	leftBytes, err := n.Left.verify()
	if err != nil {
		return nil, err
	}
	rightBytes, err := n.Right.verify()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (n *Node[T]) CalculateHash() ([]byte, error) {
	if n.leaf {
		return n.Value.Hash()
	}

	// Not sure about this check calculateNodeHash
	leftBytes, err := n.Left.CalculateHash()
	if err != nil {
		return nil, err
	}
	rightBytes, err := n.Right.CalculateHash()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// calculateNodeHash is a helper function that calculates the hash of the node.
// func (n *Node) calculateNodeHash(sort bool) ([]byte, error) {
// 	if n.leaf {
// 		return n.C.CalculateHash()
// 	}

// 	h := n.Tree.hashStrategy()
// 	if _, err := h.Write(sortAppend(sort, n.Left.Hash, n.Right.Hash)); err != nil {
// 		return nil, err
// 	}

// 	return h.Sum(nil), nil
// }

// GetMerklePath: Get Merkle path and indexes(left leaf or right leaf)
func (t *Tree[T]) GetMerklePath() ([][]byte, []int64, error) {
	for _, current := range t.Leafs {
		if current.dup {
			continue
		}

		// ok, err := current.Value.Equals()
		// // ok, err := current.Value.Equals(t.Leafs)
		// if err != nil {
		// 	return nil, nil, err
		// }

		// if ok {
		currentParent := current.Parent
		var merklePath [][]byte
		var index []int64
		for currentParent != nil {
			if bytes.Equal(currentParent.Left.Hash, current.Hash) {
				merklePath = append(merklePath, currentParent.Right.Hash)
				index = append(index, 1) // right leaf
			} else {
				merklePath = append(merklePath, currentParent.Left.Hash)
				index = append(index, 0) // left leaf
			}
			current = currentParent
			currentParent = currentParent.Parent
		}
		return merklePath, index, nil
		// }
	}
	return nil, nil, nil
}

// buildWithContent is a helper function that for a given set of Contents, generates a
// corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
// Returns an error if cs contains no Contents.
// func buildWithContent(cs []Content, t *MerkleTree) (*Node, []*Node, error) {
// 	if len(cs) == 0 {
// 		return nil, nil, errors.New("error: cannot construct tree with no content")
// 	}
// 	var leafs []*Node[T]
// 	for _, c := range cs {
// 		hash, err := c.CalculateHash()
// 		if err != nil {
// 			return nil, nil, err
// 		}

// 		leafs = append(leafs, &Node{
// 			Hash: hash,
// 			C:    c,
// 			leaf: true,
// 			Tree: t,
// 		})
// 	}
// 	if len(leafs)%2 == 1 {
// 		duplicate := &Node{
// 			Hash: leafs[len(leafs)-1].Hash,
// 			C:    leafs[len(leafs)-1].C,
// 			leaf: true,
// 			dup:  true,
// 			Tree: t,
// 		}
// 		leafs = append(leafs, duplicate)
// 	}
// 	root, err := buildIntermediate(leafs, t)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	return root, leafs, nil
// }

// buildIntermediate is a helper function that for a given list of leaf nodes, constructs
// the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate[T Hashable[T]](nl []*Node[T], t *Tree[T]) (*Node[T], error) {
	var nodes []*Node[T]
	for i := 0; i < len(nl); i += 2 {
		h := t.hashStrategy()
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}
		// chash := sortAppend(t.sort, nl[left].Hash, nl[right].Hash)
		// if _, err := h.Write(chash); err != nil {
		// 	return nil, err
		// }
		n := &Node[T]{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if len(nl) == 2 {
			return n, nil
		}
	}
	return buildIntermediate(nodes, t)
}
