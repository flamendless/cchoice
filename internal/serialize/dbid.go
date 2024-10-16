package serialize

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

//TODO: (Brandon) - implement prefix + id for safety

const PADDING = "="

func EncDBID(dbid int64) string {
	idb := make([]byte, 8)
	binary.LittleEndian.PutUint64(idb, uint64(dbid))
	id := base64.URLEncoding.EncodeToString(idb)
	id = strings.TrimSuffix(id, PADDING)
	return id
}

func DecDBID(dbid string) int64 {
	dbid += PADDING
	decid, err := base64.URLEncoding.DecodeString(dbid)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode string: '%s' %s", dbid, err.Error()))
	}
	id := int64(binary.LittleEndian.Uint64([]byte(decid)))
	return id
}
