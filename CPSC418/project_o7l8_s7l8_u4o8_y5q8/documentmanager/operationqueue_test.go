package documentmanager

import (
	. "../common"
	. "../version"
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

func NewTestVector(a, b, c uint32) VersionVector {
	v := NewVector()
	v.IncrementTo(A_ID, a)
	v.IncrementTo(B_ID, b)
	v.IncrementTo(C_ID, c)
	return v
}

func TestEnqueue1(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 1, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 2, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(5, 0, 0), Id: B_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 4, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 4, len(result))
	assertEqual(t, 1, q.Size())

	assertEqual(t, A_ID, result[0].Id)
	assertEqual(t, A_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, A_ID, result[3].Id)
	assertEqual(t, uint32(1), result[0].Version)
	assertEqual(t, uint32(2), result[1].Version)
	assertEqual(t, uint32(3), result[2].Version)
	assertEqual(t, uint32(4), result[3].Version)
}

func TestEnqueue2(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 5, len(result))
	assertEqual(t, 0, q.Size())

	assertEqual(t, A_ID, result[0].Id)
	assertEqual(t, A_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, A_ID, result[3].Id)
	assertEqual(t, B_ID, result[4].Id)
	assertEqual(t, uint32(1), result[0].Version)
	assertEqual(t, uint32(2), result[1].Version)
	assertEqual(t, uint32(3), result[2].Version)
	assertEqual(t, uint32(4), result[3].Version)
	assertEqual(t, uint32(1), result[4].Version)
}

func TestEnqueue3(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 1, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2}, NewTestVector(1, 0, 0))
	assertEqual(t, 1, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3}, NewTestVector(2, 0, 0))
	assertEqual(t, 1, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4}, NewTestVector(3, 0, 0))
	assertEqual(t, 1, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1}, NewTestVector(4, 0, 0))
	assertEqual(t, 0, q.Size())
}

func TestEnqueue4(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(4, 2, 1), Id: A_ID, Version: 5}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 2, 0), Id: C_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 5, len(result))
	assertEqual(t, 2, q.Size())

	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: B_ID, Version: 2}, NewTestVector(4, 1, 0))
	assertEqual(t, 3, len(result))
	assertEqual(t, 0, q.Size())

	assertEqual(t, B_ID, result[0].Id)
	assertEqual(t, C_ID, result[1].Id)
	assertEqual(t, A_ID, result[2].Id)
	assertEqual(t, uint32(2), result[0].Version)
	assertEqual(t, uint32(1), result[1].Version)
	assertEqual(t, uint32(5), result[2].Version)
}

func TestEnqueueDuplicateOperations(t *testing.T) {
	q := NewQueue()
	result := q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 0), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 1, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 1), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 2, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 0, 2), Id: A_ID, Version: 2}, NewTestVector(0, 0, 0))
	assertEqual(t, 0, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: C_ID, Version: 1}, NewTestVector(0, 0, 0))
	assertEqual(t, 1, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 1), Id: C_ID, Version: 2}, NewTestVector(0, 0, 1))
	assertEqual(t, 1, len(result))
	assertEqual(t, 3, q.Size())
	result = q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 2))
	assertEqual(t, 2, len(result))
	assertEqual(t, 0, q.Size())
}

func TestGetMissingQueueElem(t *testing.T) {
	q := NewQueue()
	q.Enqueue(RemoteOperation{Vector: NewTestVector(0, 0, 0), Id: A_ID, Version: 1}, NewTestVector(0, 0, 0))
	q.Enqueue(RemoteOperation{Vector: NewTestVector(1, 1, 0), Id: B_ID, Version: 2}, NewTestVector(1, 0, 0))
	q.Enqueue(RemoteOperation{Vector: NewTestVector(2, 0, 0), Id: A_ID, Version: 3}, NewTestVector(1, 0, 0))
	q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 0, 0), Id: A_ID, Version: 4}, NewTestVector(1, 0, 0))
	q.Enqueue(RemoteOperation{Vector: NewTestVector(3, 2, 0), Id: C_ID, Version: 1}, NewTestVector(1, 0, 0))
	assertEqual(t, 4, q.Size())

	result := q.GetMissingOperations(NewTestVector(2, 2, 0))
	assertEqual(t, 3, len(result))
	var a, b, c int
	for _, elem := range result {
		if EqualSiteId(elem.Id, A_ID) {
			a++
		}
		if EqualSiteId(elem.Id, B_ID) {
			b++
		}
		if EqualSiteId(elem.Id, C_ID) {
			c++
		}
	}
	assertEqual(t, 2, a)
	assertEqual(t, 0, b)
	assertEqual(t, 1, c)
}
