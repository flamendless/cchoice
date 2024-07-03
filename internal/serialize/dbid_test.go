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
		t.Fatalf("enc/dec")
	}
}
