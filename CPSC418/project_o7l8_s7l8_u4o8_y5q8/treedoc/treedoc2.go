package treedoc

import (
	"../buffer"
	"math"
)

const MAX_ATOMS_PER_NODE = math.MaxUint16

type Document struct {
	Size  int
	Doc   []*DocNode
	Nodes map[NodeId]*DocNode
}

type DocNode struct {
	Parent  *DocNode
	ParentN uint16
	NodeId  NodeId
	Size    int
	Atoms   []Atom
}

const UNINITIALIZED byte = 0
const DEAD byte = 1
const ALIVE byte = 2

type Atom struct {
	State byte
	Atom  byte
	Size  int
	Left  []*DocNode
}

const NO_OPERATION byte = 0
const INSERT_NEW byte = 1
const INSERT_ROOT byte = 2
const INSERT byte = 3
const DELETE byte = 4

type Operation struct {
	Type     byte
	ParentId NodeId
	ParentN  uint16
	Id       NodeId
	N        uint16
	Atom     byte
}

func NewDocument() *Document {
	return &Document{Doc: make([]*DocNode, 0, 4), Nodes: make(map[NodeId]*DocNode)}
}

// ***************************************************************************************
// ******************** Operations On Treedoc (Remote operations) ************************
// ***************************************************************************************

func (doc *Document) ApplyOperation(operation Operation) buffer.BufferOperation {
	if operation.Type == INSERT_NEW {
		return doc.InsertNew(operation)
	} else if operation.Type == INSERT {
		return doc.Insert(operation)
	} else if operation.Type == DELETE {
		return doc.Delete(operation)
	} else if operation.Type == INSERT_ROOT {
		return doc.InsertRoot(operation)
	}
	return buffer.BufferOperation{Type: buffer.NO_OPERATION}
}

func (doc *Document) InsertNew(operation Operation) buffer.BufferOperation {
	parent := doc.Nodes[operation.ParentId]
	newNode := &DocNode{
		NodeId:  operation.Id,
		Parent:  parent,
		ParentN: operation.ParentN,
	}
	newNode.Atoms = insertAtom(newNode.Atoms, Atom{
		State: ALIVE,
		Atom:  operation.Atom,
		Size:  1,
	}, operation.N)
	doc.Nodes[operation.Id] = newNode

	parent.Atoms = extendAtomToSize(parent.Atoms, operation.ParentN)
	atom := parent.Atoms[operation.ParentN]
	parent.Atoms[operation.ParentN] = Atom{
		State: atom.State,
		Atom:  atom.Atom,
		Size:  atom.Size,
		Left:  insertNodeIntoDisambiguatorsSorted(atom.Left, newNode),
	}
	updateSize(doc, newNode, 1)
	pos := calcPos(doc, newNode, int(operation.N))
	return buffer.BufferOperation{Type: buffer.REMOTE_INSERT, Pos: pos, Atom: operation.Atom}
}

func (doc *Document) InsertRoot(operation Operation) buffer.BufferOperation {
	newNode := &DocNode{
		Parent: nil,
		NodeId: operation.Id,
	}
	doc.Nodes[operation.Id] = newNode
	newNode.Atoms = extendAtomToSize(newNode.Atoms, operation.N)
	newNode.Atoms[operation.N] = Atom{Atom: operation.Atom, State: ALIVE, Size: 1}
	doc.Doc = insertNodeIntoDisambiguatorsSorted(doc.Doc, newNode)
	updateSize(doc, newNode, 1)
	pos := calcPos(doc, newNode, int(operation.N))
	return buffer.BufferOperation{Type: buffer.REMOTE_INSERT, Pos: pos, Atom: operation.Atom}
}

func (doc *Document) Insert(operation Operation) buffer.BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtomToSize(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State != UNINITIALIZED {
		panic("Atom is not uninitialized \"" + DocToString(doc) + "\" ")
	}
	node.Atoms[operation.N] = Atom{State: ALIVE, Atom: operation.Atom, Left: atom.Left, Size: atom.Size + 1}
	updateSize(doc, node, 1)
	pos := calcPos(doc, node, int(operation.N))
	return buffer.BufferOperation{Type: buffer.REMOTE_INSERT, Pos: pos, Atom: operation.Atom}
}

func (doc *Document) Delete(operation Operation) buffer.BufferOperation {
	node := doc.Nodes[operation.Id]
	node.Atoms = extendAtomToSize(node.Atoms, operation.N)
	atom := node.Atoms[operation.N]
	if atom.State == UNINITIALIZED {
		panic("Atom is not alive \"" + DocToString(doc) + "\" ")
	}
	if atom.State == ALIVE {
		pos := calcPos(doc, node, int(operation.N))
		node.Atoms[operation.N] = Atom{State: DEAD, Atom: atom.Atom, Left: atom.Left, Size: atom.Size - 1}
		updateSize(doc, node, -1)
		return buffer.BufferOperation{Type: buffer.DELETE, Pos: pos}
	}
	return buffer.BufferOperation{Type: buffer.NO_OPERATION}
}

// ***************************************************************************************
// ******************** Operations From buffer (Local operations) ************************
// ***************************************************************************************

func nextNonEmptyNode(i int, nodes []*DocNode) (int, *DocNode) {
	for i = i + 1; i < len(nodes); i++ {
		if nodes[i].Size != 0 {
			return i, nodes[i]
		}
	}
	return i, nil
}

func nextNonEmptyAtom(i int, atoms []Atom) int {
	for i = i + 1; i < len(atoms); i++ {
		if atoms[i].Size != 0 {
			return i
		}
	}
	return i
}

func nextNonEmptyOrUninitializedAtom(i int, atoms []Atom) int {
	for i = i + 1; i < len(atoms); i++ {
		if atoms[i].Size != 0 || atoms[i].State == UNINITIALIZED {
			return i
		}
	}
	return i
}

func findNodePos(pos int, acc int, nodes []*DocNode) (int, *DocNode) {
	i, node := nextNonEmptyNode(-1, nodes)
	for i < len(nodes) {
		if acc+node.Size > pos {
			return acc, node
		}
		acc += node.Size
		i, node = nextNonEmptyNode(i, nodes)
	}
	return acc, nil
}

func findAtomPos(pos int, acc int, atoms []Atom) (int, int) {
	i := nextNonEmptyAtom(-1, atoms)
	for i < len(atoms) {
		if acc+atoms[i].Size > pos {
			return acc, i
		}
		acc += atoms[i].Size
		i = nextNonEmptyAtom(i, atoms)
	}
	return acc, i
}

// finds the next node to keep traverse down the tree to find pos, OR find the node and atom
// where we can immediately insert atom if possible
func findAtomForInsertHelper(pos int, acc int, nodeId NodeId, node *DocNode, atoms []Atom) (byte, int, int) {
	i := nextNonEmptyAtom(-1, atoms)
	potentialInsert := EqualSiteIdInNodeId(nodeId, node.NodeId)
	for i < len(atoms) {
		if potentialInsert && acc == pos {
			if canDoInsert(node, i, nodeId) {
				return INSERT, acc, i
			} else {
				k := nextNonEmptyOrUninitializedAtom(i, atoms)
				if k < len(atoms) && canDoInsert(node, k, nodeId) {
					return INSERT, acc, k
				}
			}
		}
		if acc+atoms[i].Size > pos {
			return NO_OPERATION, acc, i
		}
		acc += atoms[i].Size
		i = nextNonEmptyAtom(i, atoms)
	}
	return NO_OPERATION, acc, i
}

// traverse the tree until it finds a place to insert (either a new node or reuse an applicable node)
func insertPosNewHelper(doc *Document, node *DocNode, n int, nodeId NodeId, ch byte) Operation {
	for {
		if len(node.Atoms[n].Left) == 0 {
			op := Operation{Type: INSERT_NEW, Id: nodeId, N: 0, ParentId: node.NodeId, ParentN: uint16(n), Atom: ch}
			doc.InsertNew(op)
			return op
		}
		node = node.Atoms[n].Left[len(node.Atoms[n].Left)-1]
		n = len(node.Atoms)
		if node.Atoms[n-1].State == UNINITIALIZED {
			n = n - 1
		}
		node.Atoms = extendAtomToSize(node.Atoms, uint16(n))
		if canDoInsert(node, n, nodeId) {
			return makeAndApplyInsertOperation(doc, node, n, ch)
		}
	}
}

func canDoInsert(node *DocNode, n int, nodeId NodeId) bool {
	return node.Atoms[n].Size == 0 && node.Atoms[n].State == UNINITIALIZED && EqualSiteIdInNodeId(nodeId, node.NodeId) && n != MAX_ATOMS_PER_NODE
}

func makeAndApplyInsertOperation(doc *Document, node *DocNode, n int, ch byte) Operation {
	op := Operation{Type: INSERT, Id: node.NodeId, N: uint16(n), Atom: ch}
	doc.Insert(op)
	return op
}

func InsertPos(doc *Document, nodeId NodeId, pos int, ch byte) Operation {
	if doc.Size == 0 {
		op := Operation{Type: INSERT_ROOT, Id: nodeId, N: 0, Atom: ch}
		doc.InsertRoot(op)
		return op
	}

	var currentN int
	var doInsert byte
	acc, currentNode := findNodePos(pos, 0, doc.Doc)
	// check for end of the document
	if currentNode == nil {
		currentNode := doc.Doc[len(doc.Doc)-1]
		currentN := len(currentNode.Atoms)
		if currentNode.Atoms[currentN-1].State == UNINITIALIZED {
			currentN = currentN - 1
		}
		currentNode.Atoms = extendAtomToSize(currentNode.Atoms, uint16(currentN))
		if canDoInsert(currentNode, currentN, nodeId) {
			return makeAndApplyInsertOperation(doc, currentNode, currentN, ch)
		}
		return insertPosNewHelper(doc, currentNode, currentN, nodeId, ch)
	}
	doInsert, acc, currentN = findAtomForInsertHelper(pos, acc, nodeId, currentNode, currentNode.Atoms)
	if doInsert == INSERT {
		return makeAndApplyInsertOperation(doc, currentNode, currentN, ch)
	}
	for {
		if acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE {
			return insertPosNewHelper(doc, currentNode, currentN, nodeId, ch)
		}
		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		doInsert, acc, currentN = findAtomForInsertHelper(pos, acc, nodeId, currentNode, currentNode.Atoms)
		if doInsert == INSERT {
			return makeAndApplyInsertOperation(doc, currentNode, currentN, ch)
		}
	}
	return Operation{Type: NO_OPERATION}
}

// traverse the tree until it finds the ALIVE atom at pos
func posToIdForDel(nodes []*DocNode, pos int) (*DocNode, int) {
	var currentN int
	acc, currentNode := findNodePos(pos, 0, nodes)
	acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	for !(acc+currentNode.Atoms[currentN].Size-1 == pos && currentNode.Atoms[currentN].State == ALIVE) {
		acc, currentNode = findNodePos(pos, acc, currentNode.Atoms[currentN].Left)
		acc, currentN = findAtomPos(pos, acc, currentNode.Atoms)
	}
	return currentNode, currentN
}

func DeletePos(doc *Document, pos int) Operation {
	node, n := posToIdForDel(doc.Doc, pos)
	op := Operation{Type: DELETE, Id: node.NodeId, N: uint16(n)}
	doc.Delete(op)
	return op
}
