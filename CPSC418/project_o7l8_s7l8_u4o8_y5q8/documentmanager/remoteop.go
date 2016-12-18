package documentmanager

import "encoding/json"
import "../treedoc"
import "../version"
import (
	. "../common"
	"github.com/nsf/termbox-go"
)

type RemoteOperation struct {
	Vector  version.VersionVector
	Id      SiteId
	Version uint32
	Op      treedoc.Operation
}

type RemoteOperationJson struct {
	Vector  version.VersionVectorJson
	Id      SiteId
	Version uint32
	Op      treedoc.Operation
}

func RemoteOperationToSlice(op RemoteOperation) []byte {
	slice, err := json.Marshal(RemoteOperationJson{
		Vector:  op.Vector.ToJsonable(),
		Id:      op.Id,
		Version: op.Version,
		Op:      op.Op,
	})
	if err != nil {
		termbox.Close()
		panic(err.Error())
	}
	return slice
}

func RemoteOperationFromSlice(slice []byte) RemoteOperation {
	var opJson RemoteOperationJson
	json.Unmarshal(slice, &opJson)
	return RemoteOperation{
		Vector:  version.FromVersionVectorJson(opJson.Vector),
		Id:      opJson.Id,
		Version: opJson.Version,
		Op:      opJson.Op,
	}
}

func RemoteOperationsToSlice(ops []RemoteOperation) []byte {
	newOps := make([]RemoteOperationJson, len(ops), len(ops))
	for i, op := range ops {
		newOps[i] = RemoteOperationJson{
			Vector:  op.Vector.ToJsonable(),
			Id:      op.Id,
			Version: op.Version,
			Op:      op.Op,
		}
	}
	b, _ := json.Marshal(newOps)
	return b
}

func RemoteOperationsFromSlice(slice []byte) []RemoteOperation {
	var opsJson []RemoteOperationJson
	json.Unmarshal(slice, &opsJson)
	ops := make([]RemoteOperation, len(opsJson), len(opsJson))
	for i, op := range opsJson {
		ops[i] = RemoteOperation{
			Vector:  version.FromVersionVectorJson(op.Vector),
			Id:      op.Id,
			Version: op.Version,
			Op:      op.Op,
		}
	}
	return ops
}
