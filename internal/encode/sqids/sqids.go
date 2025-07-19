package sqids

import (
	"cchoice/internal/encode"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"os"

	sg "github.com/sqids/sqids-go"
	"go.uber.org/zap"
)

type Sqids struct {
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
		sqids: s,
	}
}

func (sqids Sqids) Encode(dbid int64) string {
	id, err := sqids.sqids.Encode([]uint64{uint64(dbid)})
	if err != nil {
		logs.Log().Warn(
			"SQIDS",
			zap.Error(errs.ERR_ENCODE),
			zap.Error(err),
		)
		return ""
	}
	return id
}

func (sqids Sqids) Decode(id string) int64 {
	ids := sqids.sqids.Decode(id)
	if len(ids) == 0 {
		logs.Log().Warn("SQIDS", zap.Error(errs.ERR_DECODE))
		return 0
	}
	return int64(ids[0])
}

var _ encode.IEncode = (*Sqids)(nil)
