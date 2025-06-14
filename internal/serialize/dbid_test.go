package serialize

import (
	"cchoice/internal/enums"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncDecDBID(t *testing.T) {
	enc := EncodeDBID(enums.DB_PREFIX_CATEGORY, int64(123))
	prefix, dec := DecodeToDBID(enc)
	require.Equal(t, enums.DB_PREFIX_CATEGORY, prefix)
	require.Equal(t, int64(123), dec)
	require.NotPanics(t, func() {
		_ = MustDecodeToDBID(enums.DB_PREFIX_CATEGORY, enc)
	})
	require.Panics(t, func() {
		_ = MustDecodeToDBID(enums.DB_PREFIX_UNDEFINED, enc)
	})
}

func TestValidEncDecDBID(t *testing.T) {
	const encID = "Q0FURUdPUlk6MTIz"
	enc := EncodeDBID(enums.DB_PREFIX_CATEGORY, int64(123))
	require.Equal(t, encID, enc)
	dbid := MustDecodeToDBID(enums.DB_PREFIX_CATEGORY, encID)
	require.Equal(t, int64(123), dbid)
}

func BenchmarkEncDecDBID(b *testing.B) {
	for b.Loop() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		test := int64(r.Uint64())
		enc := EncodeDBID(enums.DB_PREFIX_CATEGORY, test)
		_, _ = DecodeToDBID(enc)
	}
}
