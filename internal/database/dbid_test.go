package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncDecDBID(t *testing.T) {
	enc := EncodeDBID(DB_PREFIX_CATEGORY, int64(123))
	prefix, dec := DecodeToDBID(enc)
	require.Equal(t, DB_PREFIX_CATEGORY, prefix)
	require.Equal(t, int64(123), dec)
	require.NotPanics(t, func() {
		_ = MustDecodeToDBID(DB_PREFIX_CATEGORY, enc)
	})
	require.Panics(t, func() {
		_ = MustDecodeToDBID(DB_PREFIX_UNDEFINED, enc)
	})
}

func TestValidEncDecDBID(t *testing.T) {
	const encID = "Q0FURUdPUlk6MTIz"
	enc := EncodeDBID(DB_PREFIX_CATEGORY, int64(123))
	require.Equal(t, encID, enc)
	dbid := MustDecodeToDBID(DB_PREFIX_CATEGORY, encID)
	require.Equal(t, int64(123), dbid)
}

func BenchmarkEncDecDBID(b *testing.B) {
	for i := range b.N {
		enc := EncodeDBID(DB_PREFIX_CATEGORY, int64(i))
		_, _ = DecodeToDBID(enc)
	}
}
