package treedoc

import "bytes"

func insertNodeIntoDisambiguatorsSorted(disambiguator []*DocNode, node *DocNode) []*DocNode {
	if disambiguator == nil {
		a := make([]*DocNode, 1, 4)
		a[0] = node
		return a
	}

	var a = 0
	for i := len(disambiguator); i >= 1; i-- {
		docNode := disambiguator[i-1]
		if bytes.Compare(node.NodeId[:], docNode.NodeId[:]) > 0 {
			a = i
			break
		}
	}

	disambiguator = append(disambiguator, node)
	copy(disambiguator[a+1:], disambiguator[a:])
	disambiguator[a] = node
	return disambiguator
}

// just ensure the atom slice has enough elements to prevent out of index error
func extendAtomToSize(atoms []Atom, i uint16) []Atom {
	if atoms == nil {
		return make([]Atom, i+1, i+5)
	}

	n := len(atoms)
	for int(i) >= n {
		atoms = append(atoms, Atom{})
		n++
	}
	return atoms
}

func insertAtom(atoms []Atom, atom Atom, i uint16) []Atom {
	atoms = extendAtomToSize(atoms, i)
	atoms[i] = atom
	return atoms
}

// update size from a child up the tree to the root node (delta is +1 or -1)
func updateSize(doc *Document, node *DocNode, delta int) {
	node.Size += delta
	if node.Parent == nil {
		doc.Size += delta
	} else {
		atom := node.Parent.Atoms[node.ParentN]
		node.Parent.Atoms[node.ParentN] = Atom{
			State: atom.State,
			Atom:  atom.Atom,
			Size:  atom.Size + delta,
			Left:  atom.Left,
		}
		updateSize(doc, node.Parent, delta)
	}
}

// calculate the position of the atom in the document given its parent node and its n
// used to translate remote operations to local buffer operations
func calcPos(doc *Document, node *DocNode, n int) int {
	return calcPosHelper(doc, node, n) + node.Atoms[n].Size - 1
}

func calcPosHelper(doc *Document, node *DocNode, n int) int {
	acc := 0
	for i := 0; i < n; i++ {
		acc += node.Atoms[i].Size
	}
	if node.Parent == nil {
		for _, rootNode := range doc.Doc {
			if rootNode == node {
				break
			}
			acc += rootNode.Size
		}
		return acc
	} else {
		for _, parentNode := range node.Parent.Atoms[node.ParentN].Left {
			if parentNode == node {
				break
			}
			acc += parentNode.Size
		}
		return acc + calcPosHelper(doc, node.Parent, int(node.ParentN))
	}
}
