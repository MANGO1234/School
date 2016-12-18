package documentmanager

import (
	"../treedoc"
	"testing"
)

func TestLogWrite(t *testing.T) {
	log := NewLog()
	log.Write(A_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 1)})
	assertEqual(t, 1, len(log.Log))
	assertEqual(t, NewTestVector(1, 0, 0), log.Vector)
	log.Write(A_ID, 2, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 2)})
	assertEqual(t, 2, len(log.Log))
	assertEqual(t, NewTestVector(2, 0, 0), log.Vector)
	log.Write(B_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(B_ID, 1)})
	assertEqual(t, 3, len(log.Log))
	assertEqual(t, NewTestVector(2, 1, 0), log.Vector)
	log.Write(C_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(C_ID, 1)})
	assertEqual(t, 4, len(log.Log))
	assertEqual(t, NewTestVector(2, 1, 1), log.Vector)
	log.Write(B_ID, 2, treedoc.Operation{Id: treedoc.NewNodeId(B_ID, 2)})
	assertEqual(t, 5, len(log.Log))
	assertEqual(t, NewTestVector(2, 2, 1), log.Vector)
	log.Write(A_ID, 3, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 3)})
	assertEqual(t, 6, len(log.Log))
	assertEqual(t, NewTestVector(3, 2, 1), log.Vector)
}

func TestGetMissingOperations(t *testing.T) {
	log := NewLog()
	log.Write(A_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 1)})
	log.Write(A_ID, 2, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 2)})
	log.Write(B_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(B_ID, 1)})
	log.Write(C_ID, 1, treedoc.Operation{Id: treedoc.NewNodeId(C_ID, 1)})
	log.Write(B_ID, 2, treedoc.Operation{Id: treedoc.NewNodeId(B_ID, 2)})
	log.Write(A_ID, 3, treedoc.Operation{Id: treedoc.NewNodeId(A_ID, 3)})

	assertEqual(t, 6, len(log.Log))

	result := log.GetMissingOperations(NewTestVector(2, 1, 0))
	assertEqual(t, 3, len(result))
	assertEqual(t, treedoc.NewNodeId(C_ID, 1), result[0].Op.Id)
	assertEqual(t, treedoc.NewNodeId(B_ID, 2), result[1].Op.Id)
	assertEqual(t, treedoc.NewNodeId(A_ID, 3), result[2].Op.Id)

	result = log.GetMissingOperations(NewTestVector(1, 3, 0))
	assertEqual(t, 3, len(result))
	assertEqual(t, treedoc.NewNodeId(A_ID, 2), result[0].Op.Id)
	assertEqual(t, treedoc.NewNodeId(C_ID, 1), result[1].Op.Id)
	assertEqual(t, treedoc.NewNodeId(A_ID, 3), result[2].Op.Id)

	result = log.GetMissingOperations(NewTestVector(0, 0, 0))
	assertEqual(t, 6, len(result))
	assertEqual(t, treedoc.NewNodeId(A_ID, 1), result[0].Op.Id)
	assertEqual(t, treedoc.NewNodeId(A_ID, 2), result[1].Op.Id)
	assertEqual(t, treedoc.NewNodeId(B_ID, 1), result[2].Op.Id)
}
