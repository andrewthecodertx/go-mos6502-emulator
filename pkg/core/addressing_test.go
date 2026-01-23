package core

import "testing"

func TestAddrImmediate(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)
	cpu.PC = 0x1000

	addr, pageCrossed := cpu.AddrImmediate()

	if addr != 0x1000 {
		t.Errorf("Expected addr=0x1000, got 0x%04X", addr)
	}
	if pageCrossed {
		t.Error("Immediate addressing should never cross page")
	}
	if cpu.PC != 0x1001 {
		t.Errorf("Expected PC=0x1001, got 0x%04X", cpu.PC)
	}
}

func TestAddrZeroPage(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)
	cpu.PC = 0x1000
	bus.memory[0x1000] = 0x42

	addr, pageCrossed := cpu.AddrZeroPage()

	if addr != 0x0042 {
		t.Errorf("Expected addr=0x0042, got 0x%04X", addr)
	}
	if pageCrossed {
		t.Error("Zero page addressing should never cross page")
	}
	if cpu.PC != 0x1001 {
		t.Errorf("Expected PC=0x1001, got 0x%04X", cpu.PC)
	}
}

func TestAddrZeroPageX(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("no wraparound", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0x20

		addr, pageCrossed := cpu.AddrZeroPageX()

		if addr != 0x0030 {
			t.Errorf("Expected addr=0x0030, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Zero page X should never cross page")
		}
	})

	t.Run("with wraparound", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0xF5 // 0xF5 + 0x10 = 0x105 -> wraps to 0x05

		addr, _ := cpu.AddrZeroPageX()

		if addr != 0x0005 {
			t.Errorf("Expected addr=0x0005 (wrapped), got 0x%04X", addr)
		}
	})
}

func TestAddrZeroPageY(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("no wraparound", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x10
		bus.memory[0x1000] = 0x20

		addr, pageCrossed := cpu.AddrZeroPageY()

		if addr != 0x0030 {
			t.Errorf("Expected addr=0x0030, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Zero page Y should never cross page")
		}
	})

	t.Run("with wraparound", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x20
		bus.memory[0x1000] = 0xF0 // 0xF0 + 0x20 = 0x110 -> wraps to 0x10

		addr, _ := cpu.AddrZeroPageY()

		if addr != 0x0010 {
			t.Errorf("Expected addr=0x0010 (wrapped), got 0x%04X", addr)
		}
	})
}

func TestAddrAbsolute(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)
	cpu.PC = 0x1000
	bus.memory[0x1000] = 0x34 // low byte
	bus.memory[0x1001] = 0x12 // high byte

	addr, pageCrossed := cpu.AddrAbsolute()

	if addr != 0x1234 {
		t.Errorf("Expected addr=0x1234, got 0x%04X", addr)
	}
	if pageCrossed {
		t.Error("Absolute addressing should never cross page")
	}
	if cpu.PC != 0x1002 {
		t.Errorf("Expected PC=0x1002, got 0x%04X", cpu.PC)
	}
}

func TestAddrAbsoluteX(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("no page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0x00 // low byte
		bus.memory[0x1001] = 0x20 // high byte

		addr, pageCrossed := cpu.AddrAbsoluteX()

		if addr != 0x2010 {
			t.Errorf("Expected addr=0x2010, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Expected no page cross")
		}
	})

	t.Run("with page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0xF5 // low byte
		bus.memory[0x1001] = 0x20 // high byte: 0x20F5 + 0x10 = 0x2105

		addr, pageCrossed := cpu.AddrAbsoluteX()

		if addr != 0x2105 {
			t.Errorf("Expected addr=0x2105, got 0x%04X", addr)
		}
		if !pageCrossed {
			t.Error("Expected page cross")
		}
	})
}

func TestAddrAbsoluteY(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("no page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x05
		bus.memory[0x1000] = 0x10 // low byte
		bus.memory[0x1001] = 0x30 // high byte

		addr, pageCrossed := cpu.AddrAbsoluteY()

		if addr != 0x3015 {
			t.Errorf("Expected addr=0x3015, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Expected no page cross")
		}
	})

	t.Run("with page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x20
		bus.memory[0x1000] = 0xF0 // low byte
		bus.memory[0x1001] = 0x30 // high byte: 0x30F0 + 0x20 = 0x3110

		addr, pageCrossed := cpu.AddrAbsoluteY()

		if addr != 0x3110 {
			t.Errorf("Expected addr=0x3110, got 0x%04X", addr)
		}
		if !pageCrossed {
			t.Error("Expected page cross")
		}
	})
}

func TestAddrIndirect(t *testing.T) {
	bus := &mockBus{}

	t.Run("normal case", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0x50 // pointer low
		bus.memory[0x1001] = 0x20 // pointer high -> pointer at 0x2050
		bus.memory[0x2050] = 0x34 // target low
		bus.memory[0x2051] = 0x12 // target high

		addr, pageCrossed := cpu.AddrIndirect()

		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Indirect addressing should never report page cross")
		}
	})

	t.Run("NMOS JMP indirect bug", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantNMOS)
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0xFF // pointer low (page boundary!)
		bus.memory[0x1001] = 0x20 // pointer high -> pointer at 0x20FF
		bus.memory[0x20FF] = 0x34 // target low
		bus.memory[0x2100] = 0x12 // target high (correct location)
		bus.memory[0x2000] = 0xAB // bug reads from here instead!

		addr, _ := cpu.AddrIndirect()

		// NMOS bug: reads high byte from 0x2000 instead of 0x2100
		if addr != 0xAB34 {
			t.Errorf("Expected addr=0xAB34 (NMOS bug), got 0x%04X", addr)
		}
	})

	t.Run("WDC65C02 JMP indirect fixed", func(t *testing.T) {
		cpu := NewBaseCPU(bus, VariantWDC65C02)
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0xFF // pointer low (page boundary!)
		bus.memory[0x1001] = 0x20 // pointer high -> pointer at 0x20FF
		bus.memory[0x20FF] = 0x34 // target low
		bus.memory[0x2100] = 0x12 // target high (correct location)

		addr, _ := cpu.AddrIndirect()

		// WDC65C02: bug is fixed, reads correctly from 0x2100
		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234 (fixed), got 0x%04X", addr)
		}
	})
}

func TestAddrIndirectX(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("normal case", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x05
		bus.memory[0x1000] = 0x40 // zero page addr: 0x40 + X(0x05) = 0x45
		bus.memory[0x0045] = 0x34 // target low
		bus.memory[0x0046] = 0x12 // target high

		addr, pageCrossed := cpu.AddrIndirectX()

		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Indirect X should never report page cross")
		}
	})

	t.Run("zero page wraparound on index", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0xF5 // 0xF5 + 0x10 = 0x105 -> wraps to 0x05
		bus.memory[0x0005] = 0x78 // target low
		bus.memory[0x0006] = 0x56 // target high

		addr, _ := cpu.AddrIndirectX()

		if addr != 0x5678 {
			t.Errorf("Expected addr=0x5678, got 0x%04X", addr)
		}
	})

	t.Run("zero page wraparound on pointer read", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x00
		bus.memory[0x1000] = 0xFF // pointer at 0xFF
		bus.memory[0x00FF] = 0xCD // target low
		bus.memory[0x0000] = 0xAB // target high (wraps to 0x00)

		addr, _ := cpu.AddrIndirectX()

		if addr != 0xABCD {
			t.Errorf("Expected addr=0xABCD (wrapped), got 0x%04X", addr)
		}
	})
}

func TestAddrIndirectY(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("no page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x10
		bus.memory[0x1000] = 0x40 // zero page addr
		bus.memory[0x0040] = 0x00 // base addr low
		bus.memory[0x0041] = 0x20 // base addr high: 0x2000 + Y(0x10) = 0x2010

		addr, pageCrossed := cpu.AddrIndirectY()

		if addr != 0x2010 {
			t.Errorf("Expected addr=0x2010, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Expected no page cross")
		}
	})

	t.Run("with page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x20
		bus.memory[0x1000] = 0x40 // zero page addr
		bus.memory[0x0040] = 0xF0 // base addr low
		bus.memory[0x0041] = 0x20 // base addr high: 0x20F0 + Y(0x20) = 0x2110

		addr, pageCrossed := cpu.AddrIndirectY()

		if addr != 0x2110 {
			t.Errorf("Expected addr=0x2110, got 0x%04X", addr)
		}
		if !pageCrossed {
			t.Error("Expected page cross")
		}
	})

	t.Run("zero page wraparound on pointer read", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.Y = 0x00
		bus.memory[0x1000] = 0xFF // pointer at 0xFF
		bus.memory[0x00FF] = 0x34 // base addr low
		bus.memory[0x0000] = 0x12 // base addr high (wraps)

		addr, _ := cpu.AddrIndirectY()

		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234 (wrapped), got 0x%04X", addr)
		}
	})
}

func TestAddrRelative(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantNMOS)

	t.Run("forward branch no page cross", func(t *testing.T) {
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0x10 // +16

		addr, pageCrossed := cpu.AddrRelative()

		// PC after reading offset = 0x1001, target = 0x1001 + 0x10 = 0x1011
		if addr != 0x1011 {
			t.Errorf("Expected addr=0x1011, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Expected no page cross")
		}
	})

	t.Run("forward branch with page cross", func(t *testing.T) {
		cpu.PC = 0x10F0
		bus.memory[0x10F0] = 0x20 // +32

		addr, pageCrossed := cpu.AddrRelative()

		// PC after reading = 0x10F1, target = 0x10F1 + 0x20 = 0x1111
		if addr != 0x1111 {
			t.Errorf("Expected addr=0x1111, got 0x%04X", addr)
		}
		if !pageCrossed {
			t.Error("Expected page cross")
		}
	})

	t.Run("backward branch no page cross", func(t *testing.T) {
		cpu.PC = 0x1050
		bus.memory[0x1050] = 0xF0 // -16 (0xF0 = 240, which is -16 in signed)

		addr, pageCrossed := cpu.AddrRelative()

		// PC after reading = 0x1051, target = 0x1051 - 16 = 0x1041
		if addr != 0x1041 {
			t.Errorf("Expected addr=0x1041, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Expected no page cross")
		}
	})

	t.Run("backward branch with page cross", func(t *testing.T) {
		cpu.PC = 0x1010
		bus.memory[0x1010] = 0xE0 // -32

		addr, pageCrossed := cpu.AddrRelative()

		// PC after reading = 0x1011, target = 0x1011 - 32 = 0x0FF1
		if addr != 0x0FF1 {
			t.Errorf("Expected addr=0x0FF1, got 0x%04X", addr)
		}
		if !pageCrossed {
			t.Error("Expected page cross")
		}
	})
}

func TestAddrZeroPageIndirect(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantWDC65C02)

	t.Run("normal case", func(t *testing.T) {
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0x40 // zero page addr
		bus.memory[0x0040] = 0x34 // target low
		bus.memory[0x0041] = 0x12 // target high

		addr, pageCrossed := cpu.AddrZeroPageIndirect()

		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Zero page indirect should never report page cross")
		}
	})

	t.Run("zero page wraparound", func(t *testing.T) {
		cpu.PC = 0x1000
		bus.memory[0x1000] = 0xFF // pointer at 0xFF
		bus.memory[0x00FF] = 0xCD // target low
		bus.memory[0x0000] = 0xAB // target high (wraps)

		addr, _ := cpu.AddrZeroPageIndirect()

		if addr != 0xABCD {
			t.Errorf("Expected addr=0xABCD (wrapped), got 0x%04X", addr)
		}
	})
}

func TestAddrAbsoluteIndexedIndirect(t *testing.T) {
	bus := &mockBus{}
	cpu := NewBaseCPU(bus, VariantWDC65C02)

	t.Run("normal case", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x02
		bus.memory[0x1000] = 0x00 // pointer low
		bus.memory[0x1001] = 0x20 // pointer high: 0x2000 + X(0x02) = 0x2002
		bus.memory[0x2002] = 0x34 // target low
		bus.memory[0x2003] = 0x12 // target high

		addr, pageCrossed := cpu.AddrAbsoluteIndexedIndirect()

		if addr != 0x1234 {
			t.Errorf("Expected addr=0x1234, got 0x%04X", addr)
		}
		if pageCrossed {
			t.Error("Absolute indexed indirect should never report page cross")
		}
	})

	t.Run("with index crossing page", func(t *testing.T) {
		cpu.PC = 0x1000
		cpu.X = 0x10
		bus.memory[0x1000] = 0xF5 // pointer low
		bus.memory[0x1001] = 0x20 // pointer high: 0x20F5 + X(0x10) = 0x2105
		bus.memory[0x2105] = 0x78 // target low
		bus.memory[0x2106] = 0x56 // target high

		addr, _ := cpu.AddrAbsoluteIndexedIndirect()

		if addr != 0x5678 {
			t.Errorf("Expected addr=0x5678, got 0x%04X", addr)
		}
	})
}
