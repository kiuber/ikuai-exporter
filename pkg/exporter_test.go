package pkg

import "testing"

func TestParseIfaceConnectNumAllowsValuesAboveInt8(t *testing.T) {
	got := parseIfaceConnectNum("180")
	if got != 180 {
		t.Fatalf("parseIfaceConnectNum(180) = %d, want 180", got)
	}
}

func TestParseIfaceConnectNumTreatsNonNumericAsZero(t *testing.T) {
	got := parseIfaceConnectNum("--")
	if got != 0 {
		t.Fatalf("parseIfaceConnectNum(--) = %d, want 0", got)
	}
}
