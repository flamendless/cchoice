package enums

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var tblAud = map[AudKind]string{
	AUD_UNDEFINED: "UNDEFINED",
	AUD_API:       "API",
	AUD_SYSTEM:    "SYSTEM",
}

func TestAudToString(t *testing.T) {
	for aud, val := range tblAud {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, aud.String(), val)
		})
	}
}

func TestParseAudEnum(t *testing.T) {
	for aud, val := range tblAud {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, aud, ParseAudEnum(val))
		})
	}
}

func BenchmarkAudToString(b *testing.B) {
	for aud := range tblAud {
		for i := 0; i < b.N; i++ {
			_ = aud.String()
		}
	}
}

func BenchmarkParseAudEnum(b *testing.B) {
	for _, val := range tblAud {
		for i := 0; i < b.N; i++ {
			_ = ParseAudEnum(val)
		}
	}
}
