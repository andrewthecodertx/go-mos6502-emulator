package core

import "testing"

// mockBus is a simple memory implementation for testing
type mockBus struct {
	memory [0x10000]byte
}

func (b *mockBus) Read(addr uint16) byte {
	return b.memory[addr]
}

func (b *mockBus) Write(addr uint16, data byte) {
	b.memory[addr] = data
}

func TestNewBaseCPU(t *testing.T) {
	bus := &mockBus{}

	t.Run("NMOS variant", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)

		if cpu.SP != 0xFD {
			t.Errorf("Expected SP=0xFD, got 0x%02X", cpu.SP)
		}
		if cpu.Status != 0x34 {
			t.Errorf("Expected Status=0x34, got 0x%02X", cpu.Status)
		}
		if cpu.Variant != VariantNMOS {
			t.Errorf("Expected VariantNMOS, got %v", cpu.Variant)
		}
		if cpu.A != 0 || cpu.X != 0 || cpu.Y != 0 {
			t.Error("Expected A, X, Y to be 0")
		}
		if cpu.PC != 0 {
			t.Errorf("Expected PC=0, got 0x%04X", cpu.PC)
		}
	})

	t.Run("WDC65C02 variant", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantWDC65C02)

		if cpu.Variant != VariantWDC65C02 {
			t.Errorf("Expected VariantWDC65C02, got %v", cpu.Variant)
		}
	})
}

func TestGetSetFlag(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	tests := []struct {
		flag byte
		name string
	}{
		{FlagCarry, "Carry"},
		{FlagZero, "Zero"},
		{FlagInterruptDisable, "InterruptDisable"},
		{FlagDecimal, "Decimal"},
		{FlagBreak, "Break"},
		{FlagUnused, "Unused"},
		{FlagOverflow, "Overflow"},
		{FlagNegative, "Negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all flags first
			cpu.Status = 0

			// Set the flag
			cpu.SetFlag(tt.flag, true)
			if !cpu.GetFlag(tt.flag) {
				t.Errorf("Expected %s flag to be set", tt.name)
			}

			// Clear the flag
			cpu.SetFlag(tt.flag, false)
			if cpu.GetFlag(tt.flag) {
				t.Errorf("Expected %s flag to be cleared", tt.name)
			}
		})
	}
}

func TestSetFlagPreservesOthers(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	// Set carry and zero flags
	cpu.Status = 0
	cpu.SetFlag(FlagCarry, true)
	cpu.SetFlag(FlagZero, true)

	// Verify both are set
	if !cpu.GetFlag(FlagCarry) || !cpu.GetFlag(FlagZero) {
		t.Error("Expected both Carry and Zero flags to be set")
	}

	// Clear carry, zero should remain
	cpu.SetFlag(FlagCarry, false)
	if cpu.GetFlag(FlagCarry) {
		t.Error("Expected Carry flag to be cleared")
	}
	if !cpu.GetFlag(FlagZero) {
		t.Error("Expected Zero flag to remain set")
	}
}

func TestPushPull(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)
	cpu.SP = 0xFF

	t.Run("Push decrements SP and writes to stack", func(t *testing.T) {
		cpu.Push(0x42)

		if cpu.SP != 0xFE {
			t.Errorf("Expected SP=0xFE after push, got 0x%02X", cpu.SP)
		}
		if bus.memory[0x01FF] != 0x42 {
			t.Errorf("Expected 0x42 at stack, got 0x%02X", bus.memory[0x01FF])
		}
	})

	t.Run("Pull increments SP and reads from stack", func(t *testing.T) {
		value := cpu.Pull()

		if cpu.SP != 0xFF {
			t.Errorf("Expected SP=0xFF after pull, got 0x%02X", cpu.SP)
		}
		if value != 0x42 {
			t.Errorf("Expected pulled value 0x42, got 0x%02X", value)
		}
	})

	t.Run("Stack wraparound", func(t *testing.T) {
		cpu.SP = 0x00
		cpu.Push(0xAB)

		if cpu.SP != 0xFF {
			t.Errorf("Expected SP=0xFF after wraparound, got 0x%02X", cpu.SP)
		}
		if bus.memory[0x0100] != 0xAB {
			t.Errorf("Expected 0xAB at 0x0100, got 0x%02X", bus.memory[0x0100])
		}
	})
}

func TestSetZN(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	tests := []struct {
		value        byte
		expectZero   bool
		expectNeg    bool
		description  string
	}{
		{0x00, true, false, "zero value"},
		{0x01, false, false, "positive value"},
		{0x7F, false, false, "max positive"},
		{0x80, false, true, "min negative"},
		{0xFF, false, true, "negative value"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			cpu.Status = 0
			cpu.SetZN(tt.value)

			if cpu.GetFlag(FlagZero) != tt.expectZero {
				t.Errorf("Zero flag: expected %v, got %v", tt.expectZero, cpu.GetFlag(FlagZero))
			}
			if cpu.GetFlag(FlagNegative) != tt.expectNeg {
				t.Errorf("Negative flag: expected %v, got %v", tt.expectNeg, cpu.GetFlag(FlagNegative))
			}
		})
	}
}

func TestReset(t *testing.T) {
	bus := &mockBus{}

	t.Run("NMOS reset", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)

		// Set reset vector
		bus.memory[0xFFFC] = 0x34
		bus.memory[0xFFFD] = 0x12

		// Dirty the registers
		cpu.A = 0xFF
		cpu.X = 0xFF
		cpu.Y = 0xFF
		cpu.SP = 0x00
		cpu.Status = 0xFF
		cpu.Halted = true

		cpu.Reset()

		if cpu.A != 0 || cpu.X != 0 || cpu.Y != 0 {
			t.Error("Expected A, X, Y to be cleared")
		}
		if cpu.SP != 0xFD {
			t.Errorf("Expected SP=0xFD, got 0x%02X", cpu.SP)
		}
		if cpu.PC != 0x1234 {
			t.Errorf("Expected PC=0x1234, got 0x%04X", cpu.PC)
		}
		if cpu.Cycles != 6 {
			t.Errorf("Expected 6 cycles for NMOS reset, got %d", cpu.Cycles)
		}
		if cpu.Halted {
			t.Error("Expected Halted to be false after reset")
		}
		// Status should have I flag and unused bit set
		if cpu.Status&FlagInterruptDisable == 0 {
			t.Error("Expected Interrupt Disable flag to be set")
		}
		if cpu.Status&FlagUnused == 0 {
			t.Error("Expected Unused flag to be set")
		}
	})

	t.Run("WDC65C02 reset", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantWDC65C02)

		bus.memory[0xFFFC] = 0x00
		bus.memory[0xFFFD] = 0x80

		cpu.Reset()

		if cpu.PC != 0x8000 {
			t.Errorf("Expected PC=0x8000, got 0x%04X", cpu.PC)
		}
		if cpu.Cycles != 7 {
			t.Errorf("Expected 7 cycles for WDC65C02 reset, got %d", cpu.Cycles)
		}
	})
}

func TestHandleNMI(t *testing.T) {
	bus := &mockBus{}

	t.Run("NMOS NMI", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)
		cpu.PC = 0x1234
		cpu.SP = 0xFF
		cpu.Status = 0x00
		cpu.SetFlag(FlagDecimal, true) // Set decimal mode
		cpu.NMIPending = true

		// Set NMI vector
		bus.memory[0xFFFA] = 0x00
		bus.memory[0xFFFB] = 0x30

		cpu.HandleNMI()

		// Check PC loaded from NMI vector
		if cpu.PC != 0x3000 {
			t.Errorf("Expected PC=0x3000, got 0x%04X", cpu.PC)
		}

		// Check stack contents (PC high, PC low, status)
		if bus.memory[0x01FF] != 0x12 {
			t.Errorf("Expected PC high 0x12 on stack, got 0x%02X", bus.memory[0x01FF])
		}
		if bus.memory[0x01FE] != 0x34 {
			t.Errorf("Expected PC low 0x34 on stack, got 0x%02X", bus.memory[0x01FE])
		}

		// Check I flag set
		if !cpu.GetFlag(FlagInterruptDisable) {
			t.Error("Expected Interrupt Disable flag to be set")
		}

		// NMOS should NOT clear decimal flag
		if !cpu.GetFlag(FlagDecimal) {
			t.Error("NMOS should preserve decimal flag on NMI")
		}

		// Check cycles and pending flag
		if cpu.Cycles != 7 {
			t.Errorf("Expected 7 cycles, got %d", cpu.Cycles)
		}
		if cpu.NMIPending {
			t.Error("Expected NMIPending to be cleared")
		}
	})

	t.Run("WDC65C02 NMI clears decimal", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantWDC65C02)
		cpu.PC = 0x1234
		cpu.SP = 0xFF
		cpu.Status = 0x00
		cpu.SetFlag(FlagDecimal, true)
		cpu.NMIPending = true

		bus.memory[0xFFFA] = 0x00
		bus.memory[0xFFFB] = 0x30

		cpu.HandleNMI()

		// WDC65C02 SHOULD clear decimal flag
		if cpu.GetFlag(FlagDecimal) {
			t.Error("WDC65C02 should clear decimal flag on NMI")
		}
	})
}

func TestHandleIRQ(t *testing.T) {
	bus := &mockBus{}

	t.Run("NMOS IRQ", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)
		cpu.PC = 0x5678
		cpu.SP = 0xFF
		cpu.Status = 0x00
		cpu.SetFlag(FlagDecimal, true)
		cpu.IRQPending = true

		// Set IRQ vector
		bus.memory[0xFFFE] = 0x00
		bus.memory[0xFFFF] = 0x40

		cpu.HandleIRQ()

		// Check PC loaded from IRQ vector
		if cpu.PC != 0x4000 {
			t.Errorf("Expected PC=0x4000, got 0x%04X", cpu.PC)
		}

		// Check stack contents
		if bus.memory[0x01FF] != 0x56 {
			t.Errorf("Expected PC high 0x56 on stack, got 0x%02X", bus.memory[0x01FF])
		}
		if bus.memory[0x01FE] != 0x78 {
			t.Errorf("Expected PC low 0x78 on stack, got 0x%02X", bus.memory[0x01FE])
		}

		// Check I flag set
		if !cpu.GetFlag(FlagInterruptDisable) {
			t.Error("Expected Interrupt Disable flag to be set")
		}

		// NMOS should NOT clear decimal flag
		if !cpu.GetFlag(FlagDecimal) {
			t.Error("NMOS should preserve decimal flag on IRQ")
		}

		// Check cycles and pending flag
		if cpu.Cycles != 7 {
			t.Errorf("Expected 7 cycles, got %d", cpu.Cycles)
		}
		if cpu.IRQPending {
			t.Error("Expected IRQPending to be cleared")
		}
	})

	t.Run("WDC65C02 IRQ clears decimal", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantWDC65C02)
		cpu.PC = 0x1234
		cpu.SP = 0xFF
		cpu.Status = 0x00
		cpu.SetFlag(FlagDecimal, true)
		cpu.IRQPending = true

		bus.memory[0xFFFE] = 0x00
		bus.memory[0xFFFF] = 0x40

		cpu.HandleIRQ()

		// WDC65C02 SHOULD clear decimal flag
		if cpu.GetFlag(FlagDecimal) {
			t.Error("WDC65C02 should clear decimal flag on IRQ")
		}
	})
}

func TestHandleReset(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	bus.memory[0xFFFC] = 0xCD
	bus.memory[0xFFFD] = 0xAB

	cpu.ResetPending = true
	cpu.HandleReset()

	if cpu.PC != 0xABCD {
		t.Errorf("Expected PC=0xABCD, got 0x%04X", cpu.PC)
	}
	if cpu.ResetPending {
		t.Error("Expected ResetPending to be cleared")
	}
}

func TestGetCycles(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	cpu.Cycles = 5
	if cpu.GetCycles() != 5 {
		t.Errorf("Expected GetCycles()=5, got %d", cpu.GetCycles())
	}
}
