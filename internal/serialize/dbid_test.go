package serialize

import (
	"math/rand"
	"testing"
	"time"
)

func TestEncDecDBID(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	test := int64(r.Uint64())
	enc := EncDBID(test)
	dec := DecDBID(enc)
	if test != dec {
		t.Fatalf("Fail: %d = %s = %d", test, enc, dec)
	}
}

func BenchmarkEncDecDBID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		test := int64(r.Uint64())
		enc := EncDBID(test)
		_ = DecDBID(enc)
	}
}
