package treedoc

import (
	"bytes"
	"fmt"
	"strconv"
)

// ***********************************************************
// *************** Miscellaneous *****************************
// ***********************************************************
func DocToBuffer(doc *Document) *bytes.Buffer {
	var buf bytes.Buffer
	return docToBufferHelper(doc.Doc, &buf)
}

func DocToString(doc *Document) string {
	return DocToBuffer(doc).String()
}

func docToBufferHelper(disambiguator []*DocNode, buf *bytes.Buffer) *bytes.Buffer {
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			buf = docToBufferHelper(atom.Left, buf)
			if atom.State == ALIVE {
				buf.WriteByte(atom.Atom)
			}
		}
	}
	return buf
}

func DebugDoc(doc *Document) {
	debugDocHelper(doc.Doc, "")
}

func debugDocHelper(disambiguator []*DocNode, indent string) {
	for _, node := range disambiguator {
		fmt.Print(indent + "  ")
		fmt.Printf("(%d)\n", node.Size)
		for _, atom := range node.Atoms {
			debugDocHelper(atom.Left, indent+"  ")
			if atom.State == ALIVE {
				fmt.Print(indent)
				fmt.Print("  ")
				fmt.Printf("%q", atom.Atom)
				fmt.Printf("(%d) ", atom.Size)
				fmt.Printf("%s\n", node.NodeId)
			} else if atom.State == DEAD {
				fmt.Print(indent)
				fmt.Print(" x")
				fmt.Printf("%q", atom.Atom)
				fmt.Printf("(%d) ", atom.Size)
				fmt.Print(" ")
				fmt.Printf("%s\n", node.NodeId)
			}
		}
	}
}

func DocHeight(doc *Document) int {
	return docHeightHelper(doc.Doc)
}

func docHeightHelper(disambiguator []*DocNode) int {
	max := -1
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			a := docHeightHelper(atom.Left) + 1
			if a > max {
				max = a
			}
		}
	}
	return max
}

func DocStat(doc *Document) (int, int) {
	return docStat(doc.Doc)
}

func docStat(disambiguator []*DocNode) (int, int) {
	alive := 0
	dead := 0
	for _, node := range disambiguator {
		for _, atom := range node.Atoms {
			a, d := docStat(atom.Left)
			alive += a
			dead += d
			if atom.State == ALIVE {
				alive++
			} else if atom.State == DEAD {
				dead++
			}
		}
	}
	return alive, dead
}

func AtomToString(atom Atom) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(int(atom.Atom)))
	buf.WriteString(" " + strconv.Itoa(int(atom.Size)))
	buf.WriteString(" " + strconv.Itoa(int(atom.State)))
	return buf.String()
}
