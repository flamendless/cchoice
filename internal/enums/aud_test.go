package enums

import (
	"testing"
)

func TestAudToString(t *testing.T) {
	undef := AUD_UNDEFINED
	api := AUD_API
	system := AUD_SYSTEM

	if undef.String() != "UNDEFINED" {
		t.Fatalf("Mismatch: %s = %s", undef.String(), "UNDEFINED")
	}

	if api.String() != "API" {
		t.Fatalf("Mismatch: %s = %s", api.String(), "API")
	}

	if system.String() != "SYSTEM" {
		t.Fatalf("Mismatch: %s = %s", system.String(), "SYSTEM")
	}
}

func TestParseAudEnum(t *testing.T) {
	undef := ParseAudEnum("UNDEFINED")
	asc := ParseAudEnum("API")
	desc := ParseAudEnum("SYSTEM")

	if undef != AUD_UNDEFINED {
		t.Fatalf("Mismatch: %s = %s", undef, AUD_UNDEFINED)
	}
	if asc != AUD_API {
		t.Fatalf("Mismatch: %s = %s", asc, AUD_API)
	}
	if desc != AUD_SYSTEM {
		t.Fatalf("Mismatch: %s = %s", desc, AUD_SYSTEM)
	}
}
