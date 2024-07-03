package serialize

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

func EncDBID(dbid int64) string {
	idb := make([]byte, 8)
	binary.LittleEndian.PutUint64(idb, uint64(dbid))
	id := base64.StdEncoding.EncodeToString(idb)
	return id
}

func DecDBID(dbid string) int64 {
	decid, err := base64.StdEncoding.DecodeString(dbid)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode string: '%s' %s", dbid, err.Error()))
	}
	id := int64(binary.LittleEndian.Uint64([]byte(decid)))
	return id
}
