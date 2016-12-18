package treedoc

import . "../common"

type NodeId [20]byte

func NewNodeId(siteId SiteId, clock uint32) NodeId {
	clockByte := Uint32ToBytes(clock)
	var ID NodeId
	copy(ID[:], siteId[:])
	copy(ID[16:], clockByte)
	return ID
}

func StringToNodeId(id string) NodeId {
	var b [20]byte
	copy(b[:], id[:])
	return b
}

func SeparateNodeID(id NodeId) (SiteId, uint32) {
	var siteId SiteId
	copy(siteId[:], id[:16])
	version := BytesToUint32(id[16:])
	return siteId, version
}

func EqualNodeId(a NodeId, b NodeId) bool {
	for i := 0; i < 20; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func EqualSiteIdInNodeId(a NodeId, b NodeId) bool {
	for i := 0; i < 16; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
