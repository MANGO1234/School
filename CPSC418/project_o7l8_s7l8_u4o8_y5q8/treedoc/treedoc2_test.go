package treedoc

import (
	"../buffer"
	"reflect"
	"runtime/debug"
	"testing"
)

func assertEqual(t *testing.T, exp, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		debug.PrintStack()
		t.Fatalf("Expecting '%v' got '%v'\n", exp, got)
	}
}

func newId(id string) NodeId {
	var ID [20]byte
	copy(ID[:], id[:])
	return ID
}

var A_ID0 = newId("aaaaaaaaaaaaaaaa0000")
var A_ID1 = newId("aaaaaaaaaaaaaaaa0001")
var A_ID2 = newId("aaaaaaaaaaaaaaaa0002")
var A_ID3 = newId("aaaaaaaaaaaaaaaa0003")
var B_ID0 = newId("bbbbbbbbbbbbbbbb0000")
var B_ID1 = newId("bbbbbbbbbbbbbbbb0001")
var C_ID0 = newId("cccccccccccccccc0000")
var C_ID1 = newId("cccccccccccccccc0001")
var D_ID0 = newId("dddddddddddddddd0000")

func NewTestDoc() *Document {
	d := NewDocument()
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'e', ParentId: C_ID1, ParentN: 1, Id: D_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT, Atom: 'f', Id: C_ID0, N: 1})
	d.ApplyOperation(Operation{Type: INSERT, Atom: 'g', Id: C_ID1, N: 1})
	d.ApplyOperation(Operation{Type: INSERT, Atom: 'h', Id: C_ID1, N: 2})
	return d
}

// basic op
func TestInsert(t *testing.T) {
	d := NewDocument()
	assertEqual(t, -1, DocHeight(d))
	assertEqual(t, 0, d.Size)

	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "ab", DocToString(d))
	assertEqual(t, 0, DocHeight(d))
	assertEqual(t, 2, d.Size)

	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'e', ParentId: C_ID1, ParentN: 1, Id: D_ID0, N: 0})
	assertEqual(t, "cadeb", DocToString(d))
	assertEqual(t, 2, DocHeight(d))
	assertEqual(t, 5, d.Size)

	d.ApplyOperation(Operation{Type: INSERT, Atom: 'f', Id: C_ID0, N: 1})
	d.ApplyOperation(Operation{Type: INSERT, Atom: 'g', Id: C_ID1, N: 1})
	d.ApplyOperation(Operation{Type: INSERT, Atom: 'h', Id: C_ID1, N: 2})
	assertEqual(t, "cfadeghb", DocToString(d))
	assertEqual(t, 2, DocHeight(d))
	assertEqual(t, 8, d.Size)

	d = NewDocument()
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'c', Id: C_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "abc", DocToString(d))
	assertEqual(t, 3, d.Size)
}

func TestInsertDisambiguator(t *testing.T) {
	d := NewDocument()
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'c', Id: C_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, "abc", DocToString(d))
	assertEqual(t, 3, d.Size)

	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: C_ID0, ParentN: 1, Id: C_ID1, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'a', ParentId: C_ID0, ParentN: 1, Id: A_ID1, N: 0})
	d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'b', ParentId: C_ID0, ParentN: 1, Id: B_ID1, N: 0})
	assertEqual(t, "abcabc", DocToString(d))
	assertEqual(t, 6, d.Size)
}

func TestDelete(t *testing.T) {
	d := NewTestDoc()
	d.ApplyOperation(Operation{Type: DELETE, Id: C_ID1, N: 1})
	d.ApplyOperation(Operation{Type: DELETE, Id: C_ID0, N: 0})
	assertEqual(t, "fadehb", DocToString(d))
	assertEqual(t, 6, d.Size)
	x, y := DocStat(d)
	assertEqual(t, 6, x)
	assertEqual(t, 2, y)
}

//commutative
func TestInsertInsert(t *testing.T) {
	d1 := NewDocument()
	d2 := NewDocument()

	d1.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	d1.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	d2.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	d2.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	assertEqual(t, DocToString(d1), "ab")
	assertEqual(t, DocToString(d2), "ab")
	assertEqual(t, 2, d1.Size)
	assertEqual(t, 2, d1.Size)

	d1.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	d1.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	d2.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	d2.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	assertEqual(t, DocToString(d1), "cadb")
	assertEqual(t, DocToString(d2), "cadb")
	assertEqual(t, 4, d1.Size)
	assertEqual(t, 4, d1.Size)

	d1.ApplyOperation(Operation{Type: INSERT, Atom: 'e', Id: C_ID1, N: 1})
	d1.ApplyOperation(Operation{Type: INSERT, Atom: 'f', Id: C_ID1, N: 2})
	d2.ApplyOperation(Operation{Type: INSERT, Atom: 'f', Id: C_ID1, N: 2})
	d2.ApplyOperation(Operation{Type: INSERT, Atom: 'e', Id: C_ID1, N: 1})
	assertEqual(t, DocToString(d1), "cadefb")
	assertEqual(t, DocToString(d2), "cadefb")
	assertEqual(t, 6, d1.Size)
	assertEqual(t, 6, d1.Size)
}

func TestDeleteDelete(t *testing.T) {
	d := NewTestDoc()
	d.ApplyOperation(Operation{Type: DELETE, Id: C_ID1, N: 1})
	assertEqual(t, "cfadehb", DocToString(d))
	x, y := DocStat(d)
	assertEqual(t, 7, x)
	assertEqual(t, 1, y)

	d.ApplyOperation(Operation{Type: DELETE, Id: C_ID1, N: 1})
	assertEqual(t, "cfadehb", DocToString(d))
	x, y = DocStat(d)
	assertEqual(t, 7, x)
	assertEqual(t, 1, y)
}

func TestDeletePos(t *testing.T) {
	for i := 0; i < 8; i++ {
		d := NewTestDoc()
		DeletePos(d, i)
		assertEqual(t, "cfadeghb"[:i]+"cfadeghb"[i+1:], DocToString(d))
		assertEqual(t, 7, d.Size)
		x, y := DocStat(d)
		assertEqual(t, 7, x)
		assertEqual(t, 1, y)
	}

	for i := 0; i < 7; i++ {
		for j := 0; j < 6; j++ {
			d := NewTestDoc()
			str := "cfadeghb"
			DeletePos(d, i)
			str = str[:i] + str[i+1:]
			assertEqual(t, str, DocToString(d))
			DeletePos(d, j)
			str = str[:j] + str[j+1:]
			assertEqual(t, str, DocToString(d))
			assertEqual(t, 6, d.Size)
			x, y := DocStat(d)
			assertEqual(t, 6, x)
			assertEqual(t, 2, y)
		}
	}

	d := NewTestDoc()
	assertEqual(t, "cfadeghb", DocToString(d))
	DeletePos(d, 0)
	assertEqual(t, "fadeghb", DocToString(d))
	assertEqual(t, 7, d.Size)
	x, y := DocStat(d)
	assertEqual(t, 7, x)
	assertEqual(t, 1, y)

	DeletePos(d, 4)
	assertEqual(t, "fadehb", DocToString(d))
	assertEqual(t, 6, d.Size)
	x, y = DocStat(d)
	assertEqual(t, 6, x)
	assertEqual(t, 2, y)
}

func TestBufferOperationReturn(t *testing.T) {
	d := NewTestDoc()
	op := d.ApplyOperation(Operation{Type: DELETE, Id: C_ID1, N: 1})
	assertEqual(t, buffer.DELETE, op.Type)
	assertEqual(t, 5, op.Pos)
	op = d.ApplyOperation(Operation{Type: DELETE, Id: C_ID0, N: 0})
	assertEqual(t, buffer.DELETE, op.Type)
	assertEqual(t, 0, op.Pos)

	d = NewDocument()
	op = d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 0, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'b', Id: B_ID0, N: 0})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', ParentId: A_ID0, ParentN: 0, Id: C_ID0, N: 0})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 0, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'd', ParentId: A_ID0, ParentN: 1, Id: C_ID1, N: 0})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 2, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'e', ParentId: C_ID1, ParentN: 1, Id: D_ID0, N: 0})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 3, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'f', Id: C_ID0, N: 1})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'g', Id: C_ID1, N: 1})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 5, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'h', Id: C_ID1, N: 2})
	assertEqual(t, buffer.REMOTE_INSERT, op.Type)
	assertEqual(t, 6, op.Pos)

	d = NewDocument()
	op = d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	assertEqual(t, 0, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 1})
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 2})
	assertEqual(t, 2, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 3})
	assertEqual(t, 3, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'b', Id: A_ID1, N: 0, ParentId: A_ID0, ParentN: 1})
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', Id: B_ID1, N: 0, ParentId: A_ID0, ParentN: 1})
	assertEqual(t, 2, op.Pos)
	assertEqual(t, "abcaaa", DocToString(d))
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'b', Id: A_ID1, N: 1})
	assertEqual(t, 2, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'c', Id: B_ID1, N: 1})
	assertEqual(t, 4, op.Pos)

	d = NewDocument()
	op = d.ApplyOperation(Operation{Type: INSERT_ROOT, Atom: 'a', Id: A_ID0, N: 0})
	assertEqual(t, 0, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 1})
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 2})
	assertEqual(t, 2, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'a', Id: A_ID0, N: 3})
	assertEqual(t, 3, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'c', Id: B_ID1, N: 0, ParentId: A_ID0, ParentN: 1})
	assertEqual(t, 1, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT_NEW, Atom: 'b', Id: A_ID1, N: 0, ParentId: A_ID0, ParentN: 1})
	assertEqual(t, 1, op.Pos)
	assertEqual(t, "abcaaa", DocToString(d))
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'b', Id: A_ID1, N: 1})
	assertEqual(t, 2, op.Pos)
	op = d.ApplyOperation(Operation{Type: INSERT, Atom: 'c', Id: B_ID1, N: 1})
	assertEqual(t, 4, op.Pos)
}

func TestInsertPos(t *testing.T) {
	for i := 0; i < 9; i++ {
		d := NewTestDoc()
		InsertPos(d, A_ID2, i, 'z')
		assertEqual(t, "cfadeghb"[:i]+"z"+"cfadeghb"[i:], DocToString(d))
		assertEqual(t, 9, d.Size)
	}

	for i := 0; i < 9; i++ {
		for j := 0; j < 10; j++ {
			d := NewTestDoc()
			str := "cfadeghb"
			InsertPos(d, A_ID2, i, 'z')
			str = str[:i] + "z" + str[i:]
			assertEqual(t, str, DocToString(d))
			InsertPos(d, A_ID3, j, 'y')
			str = str[:j] + "y" + str[j:]
			assertEqual(t, str, DocToString(d))
			assertEqual(t, 10, d.Size)
		}
	}
}

func TestInsertPosOpReturn(t *testing.T) {
	d := NewDocument()
	assertEqual(t, -1, DocHeight(d))

	op := InsertPos(d, A_ID0, 0, 'z')
	assertEqual(t, "z", DocToString(d))
	assertEqual(t, INSERT_ROOT, op.Type)
	assertEqual(t, 1, d.Size)
	assertEqual(t, 0, DocHeight(d))

	op = InsertPos(d, A_ID1, 1, 'y')
	assertEqual(t, "zy", DocToString(d))
	assertEqual(t, INSERT, op.Type)
	assertEqual(t, 2, d.Size)
	assertEqual(t, 0, DocHeight(d))

	op = InsertPos(d, A_ID1, 2, 'x')
	assertEqual(t, "zyx", DocToString(d))
	assertEqual(t, INSERT, op.Type)
	assertEqual(t, 3, d.Size)
	assertEqual(t, 0, DocHeight(d))

	op = InsertPos(d, A_ID1, 2, 'a')
	assertEqual(t, "zyax", DocToString(d))
	assertEqual(t, INSERT_NEW, op.Type)
	assertEqual(t, 4, d.Size)
	assertEqual(t, 1, DocHeight(d))

	op = InsertPos(d, A_ID2, 3, 'b')
	assertEqual(t, "zyabx", DocToString(d))
	assertEqual(t, INSERT, op.Type)
	assertEqual(t, 5, d.Size)
	assertEqual(t, 1, DocHeight(d))

	op = InsertPos(d, A_ID2, 0, 'c')
	assertEqual(t, "czyabx", DocToString(d))
	assertEqual(t, INSERT_NEW, op.Type)
	assertEqual(t, 6, d.Size)
	assertEqual(t, 1, DocHeight(d))

	op = InsertPos(d, A_ID2, 3, 'd')
	assertEqual(t, "czydabx", DocToString(d))
	assertEqual(t, INSERT_NEW, op.Type)
	assertEqual(t, 7, d.Size)
	assertEqual(t, 2, DocHeight(d))

	d = NewDocument()
	op = InsertPos(d, A_ID0, 0, 'z')
	op = InsertPos(d, A_ID0, 1, 'z')
	op = InsertPos(d, A_ID0, 2, 'z')
	op = InsertPos(d, A_ID0, 3, 'z')
	op = InsertPos(d, A_ID1, 1, 'y')
	op = InsertPos(d, B_ID1, 1, 'x')
	assertEqual(t, 2, DocHeight(d))
	assertEqual(t, "zxyzzz", DocToString(d))

	d = NewDocument()
	op = InsertPos(d, A_ID0, 0, 'z')
	op = InsertPos(d, A_ID0, 1, 'z')
	op = InsertPos(d, A_ID0, 2, 'z')
	op = InsertPos(d, A_ID0, 3, 'z')
	op = InsertPos(d, B_ID1, 1, 'x')
	op = InsertPos(d, A_ID1, 1, 'y')
	assertEqual(t, 2, DocHeight(d))
	assertEqual(t, "zyxzzz", DocToString(d))
}

// quadratic insert in this case -> will fix if have time
// temporary reduce MAX_ATOMS_PER_NODE to MaxUint8 will work
//func TestInsertPosMaximumCharacters(t *testing.T) {
//	d := NewDocument()
//	op := InsertPos(d, A_ID0, 0, 'z')
//	assertEqual(t, "z", DocToString(d))
//	assertEqual(t, INSERT_ROOT, op.Type)
//	assertEqual(t, 1, d.Size)
//	assertEqual(t, 0, DocHeight(d))
//
//	for i := 1; i < MAX_ATOMS_PER_NODE; i++ {
//		op := InsertPos(d, A_ID1, i, 'z')
//		assertEqual(t, INSERT, op.Type)
//		assertEqual(t, i+1, d.Size)
//		assertEqual(t, 0, DocHeight(d))
//	}
//
//	var buf bytes.Buffer
//	for i := 0; i < MAX_ATOMS_PER_NODE; i++ {
//		buf.WriteByte('z')
//	}
//	assertEqual(t, buf.String(), DocToString(d))
//
//	op = InsertPos(d, A_ID1, MAX_ATOMS_PER_NODE, 'z')
//	buf.WriteByte('z')
//	assertEqual(t, buf.String(), DocToString(d))
//	assertEqual(t, INSERT_NEW, op.Type)
//	assertEqual(t, MAX_ATOMS_PER_NODE+1, d.Size)
//	assertEqual(t, 1, DocHeight(d))
//}
