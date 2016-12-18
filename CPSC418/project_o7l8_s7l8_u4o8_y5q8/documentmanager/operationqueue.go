package documentmanager

import (
	. "../version"
)

type OperationQueue struct {
	queue  []RemoteOperation
	vector VersionVector
}

func NewQueue() *OperationQueue {
	return &OperationQueue{make([]RemoteOperation, 0, 4), NewVector()}
}

func (queue *OperationQueue) Size() int {
	return len(queue.queue)
}

// enqueue an operation and returns list of operation that's ready
func (queue *OperationQueue) Enqueue(elem RemoteOperation, vector VersionVector) []RemoteOperation {
	if vector.Get(elem.Id) == elem.Version-1 {
		compare := vector.Compare(elem.Vector)
		if compare == GREATER_THAN || compare == EQUAL {
			result := make([]RemoteOperation, 1, 4)
			vector.Increment(elem.Id)
			result[0] = elem
			result, offset := dequeHelper(result, queue, len(queue.queue), vector)
			queue.queue = queue.queue[:len(queue.queue)-offset]
			return result
		}
	}
	// one assumption we made is operations from a site are received in order
	//if queue.vector.Get(elem.Id) < elem.Version {
	queue.queue = append(queue.queue, elem)
	//queue.vector.IncrementTo(elem.Id, elem.Version)
	//}
	return nil
}

func dequeHelper(result []RemoteOperation, queue *OperationQueue, upto int, vector VersionVector) ([]RemoteOperation, int) {
	q := queue.queue
	v := vector
	offset := 0
	for i := 0; i < upto; i++ {
		if offset != 0 {
			q[i-offset] = q[i]
		}
		// remove operation that already exists
		if v.Get(q[i].Id) >= q[i].Version {
			offset += 1
			continue
		}
		compare := v.Compare(q[i].Vector)
		if compare == GREATER_THAN || compare == EQUAL {
			v.Increment(q[i].Id)
			result = append(result, q[i])
			var offsetp int
			result, offsetp = dequeHelper(result, queue, i-offset, vector)
			offset += offsetp + 1
		}
	}
	return result, offset
}

func (queue *OperationQueue) GetMissingOperations(vector VersionVector) []RemoteOperation {
	q := queue.queue
	result := make([]RemoteOperation, 0)
	for i := 0; i < len(q); i++ {
		elem := q[i]
		if vector.Get(elem.Id) < elem.Version {
			result = append(result, elem)
		}
	}
	return result
}
