package core

import "testing"

func TestVariantString(t *testing.T) {
	tests := []struct {
		variant  Variant
		expected string
	}{
		{VariantNMOS, "NMOS6502"},
		{VariantWDC65C02, "WDC65C02"},
		{Variant(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.variant.String(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestVariantResetCycles(t *testing.T) {
	tests := []struct {
		variant  Variant
		expected byte
	}{
		{VariantNMOS, 6},
		{VariantWDC65C02, 7},
		{Variant(99), 6}, // default
	}

	for _, tt := range tests {
		t.Run(tt.variant.String(), func(t *testing.T) {
			if got := tt.variant.ResetCycles(); got != tt.expected {
				t.Errorf("Expected %d cycles, got %d", tt.expected, got)
			}
		})
	}
}

func TestVariantHasJMPIndirectBug(t *testing.T) {
	if !VariantNMOS.HasJMPIndirectBug() {
		t.Error("NMOS should have JMP indirect bug")
	}
	if VariantWDC65C02.HasJMPIndirectBug() {
		t.Error("WDC65C02 should NOT have JMP indirect bug")
	}
}

func TestVariantClearsDecimalOnInterrupt(t *testing.T) {
	if VariantNMOS.ClearsDecimalOnInterrupt() {
		t.Error("NMOS should NOT clear decimal on interrupt")
	}
	if !VariantWDC65C02.ClearsDecimalOnInterrupt() {
		t.Error("WDC65C02 should clear decimal on interrupt")
	}
}
