package enums

import (
	"testing"
)

var tblAud = map[AudKind]string{
	AUD_UNDEFINED: "UNDEFINED",
	AUD_API:       "API",
	AUD_SYSTEM:    "SYSTEM",
}

func TestAudToString(t *testing.T) {
	for aud, val := range tblAud {
		if aud.String() != val {
			t.Fatalf("Mismatch: %s = %s", aud.String(), val)
		}
	}
}

func TestParseAudEnum(t *testing.T) {
	for aud, val := range tblAud {
		parsed := ParseAudEnum(val)
		if parsed != aud {
			t.Fatalf("Mismatch: %s = %s", val, aud)
		}
	}
}
