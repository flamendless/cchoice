package sqids

import (
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"errors"
	"os"

	sg "github.com/sqids/sqids-go"
	"go.uber.org/zap"
)

type Sqids struct {
	name  string
	sqids *sg.Sqids
}

func MustSqids() *Sqids {
	alphabet := os.Getenv("ENCODE_SALT")
	if alphabet == "" {
		panic(errs.ERR_ENV_VAR_REQUIRED)
	}

	s, err := sg.New(sg.Options{
		MinLength: 16,
		Alphabet:  alphabet,
	})
	if err != nil {
		panic(err)
	}
	return &Sqids{
		name:  "SQIDS",
		sqids: s,
	}
}

func (sqids Sqids) Name() string {
	return sqids.name
}

func (sqids Sqids) Encode(dbid int64) string {
	id, err := sqids.sqids.Encode([]uint64{uint64(dbid)})
	if err != nil {
		logs.Log().Warn(
			sqids.Name(),
			zap.Error(errors.Join(errs.ERR_ENCODE, err)),
			zap.Int64("dbid", dbid),
		)
		return ""
	}
	return id
}

func (sqids Sqids) Decode(id string) int64 {
	ids := sqids.sqids.Decode(id)
	if len(ids) == 0 {
		logs.Log().Warn(
			sqids.Name(),
			zap.Error(errs.ERR_DECODE),
			zap.String("id", id),
		)
		return -1
	}
	return int64(ids[0])
}

var _ encode.IEncode = (*Sqids)(nil)
