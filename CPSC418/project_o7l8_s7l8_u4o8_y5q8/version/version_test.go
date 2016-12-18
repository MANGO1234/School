package version

import (
	. "../common"
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

var A_ID = StringToSiteId("aaaaaaaaaaaaaaaa")
var B_ID = StringToSiteId("bbbbbbbbbbbbbbbb")
var C_ID = StringToSiteId("cccccccccccccccc")

func TestIncrement(t *testing.T) {
	v := NewVector()
	v.Increment(A_ID)
	assertEqual(t, uint32(1), v.Get(A_ID))
}

func TestIncrementTo(t *testing.T) {
	v := NewVector()
	v.IncrementTo(A_ID, 5)
	assertEqual(t, uint32(5), v.Get(A_ID))
	v.IncrementTo(A_ID, 2)
	assertEqual(t, uint32(5), v.Get(A_ID))
}

func TestMerge(t *testing.T) {
	v1 := NewVector()
	v1.IncrementTo(A_ID, 5)
	v1.IncrementTo(B_ID, 2)
	v1.IncrementTo(C_ID, 1)
	vc := v1.Copy()

	v2 := NewVector()
	v2.IncrementTo(A_ID, 3)
	v2.IncrementTo(B_ID, 6)

	v1.Merge(v2)
	assertEqual(t, uint32(5), v1.Get(A_ID))
	assertEqual(t, uint32(6), v1.Get(B_ID))
	assertEqual(t, uint32(1), v1.Get(C_ID))

	v2.Merge(vc)
	assertEqual(t, uint32(5), v2.Get(A_ID))
	assertEqual(t, uint32(6), v2.Get(B_ID))
	assertEqual(t, uint32(1), v2.Get(C_ID))
}

func TestCompare(t *testing.T) {
	v1 := NewVector()
	v1.IncrementTo(A_ID, 5)
	v1.IncrementTo(B_ID, 2)
	v1.IncrementTo(C_ID, 1)
	v2 := NewVector()
	v2.IncrementTo(A_ID, 5)
	v2.IncrementTo(B_ID, 2)
	v2.IncrementTo(C_ID, 1)
	assertEqual(t, EQUAL, v1.Compare(v2))
	assertEqual(t, EQUAL, v2.Compare(v1))

	v1 = NewVector()
	v1.IncrementTo(A_ID, 5)
	v1.IncrementTo(B_ID, 2)
	v1.IncrementTo(C_ID, 1)
	v2 = NewVector()
	v2.IncrementTo(A_ID, 5)
	v2.IncrementTo(B_ID, 2)
	assertEqual(t, GREATER_THAN, v1.Compare(v2))
	assertEqual(t, LESS_THAN, v2.Compare(v1))

	v1 = NewVector()
	v1.IncrementTo(A_ID, 5)
	v1.IncrementTo(B_ID, 2)
	v2 = NewVector()
	v2.IncrementTo(A_ID, 5)
	v2.IncrementTo(B_ID, 2)
	v2.IncrementTo(C_ID, 1)
	assertEqual(t, LESS_THAN, v1.Compare(v2))
	assertEqual(t, GREATER_THAN, v2.Compare(v1))

	v1 = NewVector()
	v1.IncrementTo(A_ID, 5)
	v1.IncrementTo(B_ID, 2)
	v1.IncrementTo(C_ID, 1)
	v2 = NewVector()
	v2.IncrementTo(A_ID, 4)
	v2.IncrementTo(B_ID, 2)
	v2.IncrementTo(C_ID, 3)
	assertEqual(t, CONFLICT, v1.Compare(v2))
	assertEqual(t, CONFLICT, v2.Compare(v1))

	v1 = NewVector()
	v1.IncrementTo(A_ID, 5)
	v2 = NewVector()
	v2.IncrementTo(C_ID, 3)
	assertEqual(t, CONFLICT, v1.Compare(v2))
	assertEqual(t, CONFLICT, v2.Compare(v1))

	v1 = NewVector()
	v1.IncrementTo(A_ID, 6)
	v1.IncrementTo(B_ID, 2)
	v2 = NewVector()
	v2.IncrementTo(B_ID, 2)
	v2.IncrementTo(C_ID, 1)
	assertEqual(t, CONFLICT, v1.Compare(v2))
	assertEqual(t, CONFLICT, v2.Compare(v1))
}
