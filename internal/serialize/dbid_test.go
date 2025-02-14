package serialize

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncDecDBID(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	test := int64(r.Uint64())
	enc := EncDBID(test)
	dec := DecDBID(enc)
	require.Equal(t, test, dec)
}

func BenchmarkEncDecDBID(b *testing.B) {
	for b.Loop() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		test := int64(r.Uint64())
		enc := EncDBID(test)
		_ = DecDBID(enc)
	}
}
