package documentmanager

import (
	. "../common"
	"../treedoc"
	"../version"
)

type LogEntry struct {
	Operation treedoc.Operation
	Id        SiteId
	Version   uint32
}

type OperationLog struct {
	Log    []LogEntry
	Vector version.VersionVector
}

func NewLog() *OperationLog {
	return &OperationLog{make([]LogEntry, 0, 100), version.NewVector()}
}

func (log *OperationLog) Write(id SiteId, version uint32, operation treedoc.Operation) {
	log.Vector.IncrementTo(id, version)
	log.Log = append(log.Log, LogEntry{
		Id:        id,
		Version:   version,
		Operation: operation,
	})
}

func (log *OperationLog) GetMissingOperations(vector version.VersionVector) []RemoteOperation {
	result := make([]RemoteOperation, 0, 10)
	emptyVector := version.NewVector()
	currentVector := vector.Copy()

	for i := len(log.Log) - 1; i >= 0; i-- {
		currentLog := log.Log[i]
		if vector.Get(currentLog.Id) < currentLog.Version {
			result = append(result, RemoteOperation{
				Vector:  emptyVector,
				Id:      currentLog.Id,
				Version: currentLog.Version,
				Op:      currentLog.Operation,
			})
		}
		currentVector.DecrementTo(currentLog.Id, currentLog.Version-1)

		compare := currentVector.Compare(vector)
		if compare == version.EQUAL && compare == version.LESS_THAN {
			break
		}
	}

	for i := len(result)/2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}
	return result
}
