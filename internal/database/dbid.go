package database

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
)

func EncodeDBID(prefix DBPrefix, dbid int64) string {
	var buf bytes.Buffer
	buf.WriteString(prefix.String())
	buf.WriteByte(':')
	buf.WriteString(strconv.FormatInt(dbid, 10))
	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func DecodeToDBID(dbid string) (DBPrefix, int64) {
	dec, err := base64.RawURLEncoding.DecodeString(dbid)
	if err != nil {
		panic(fmt.Errorf("Failed to decode string: '%s' %w", dec, err))
	}

	splitted := bytes.Split(dec, []byte(":"))
	if len(splitted) != 2 {
		panic(fmt.Errorf("Failed to split bytestring: '%+v'", splitted))
	}

	prefix, decid := splitted[0], splitted[1]

	id, err := strconv.Atoi(string(decid))
	if err != nil {
		panic(fmt.Errorf("Failed to decode string: '%s' %w", decid, err))
	}

	return ParseDBPrefixToEnum(string(prefix)), int64(id)
}

func MustDecodeToDBID(prefix DBPrefix, dbid string) int64 {
	dec, err := base64.RawURLEncoding.DecodeString(dbid)
	if err != nil {
		panic(fmt.Errorf("Failed to decode string '%s' from '%s' %w", dec, dbid, err))
	}

	splitted := bytes.Split(dec, []byte(":"))
	if len(splitted) != 2 {
		panic(fmt.Errorf("Failed to split bytestring: '%+v'", splitted))
	}

	if string(splitted[0]) != prefix.String() {
		panic(fmt.Errorf("Prefix must match: '%s' %s", prefix.String(), splitted[0]))
	}

	id, err := strconv.Atoi(string(splitted[1]))
	if err != nil {
		panic(fmt.Errorf("Failed to decode string: '%s' %w", splitted[1], err))
	}

	return int64(id)
}
