# go-6502-emulator

A cycle-accurate emulator for the MOS6502 and WDC65C02 8-bit microprocessors,
written in Go.

## Overview

This project provides high-fidelity emulation of two classic 8-bit processors
that powered many iconic computers and gaming systems:

- **MOS6502**: The original MOS Technology 6502 processor (1975)
- **WDC65C02**: The Western Design Center's enhanced CMOS version (1982)

## Features

### Core Capabilities

- **Cycle-accurate execution**: Faithfully reproduces timing behavior including
page-crossing penalties
- **Complete instruction set**: All 56 documented MOS6502 instructions plus 27
additional WDC65C02 instructions
- **All addressing modes**: 13 addressing modes for 6502, plus 2 new modes for
65C02
- **Hardware quirks**: Implements the JMP indirect page boundary bug on MOS6502
(fixed in 65C02)
- **Interrupt support**: NMI, IRQ, and RESET vectors with proper handling
- **Flexible bus interface**: Pluggable memory system supporting RAM, ROM, and
memory-mapped I/O

### MOS6502 Features

- 56 documented instructions
- JMP indirect page boundary bug (hardware-accurate)
- 6-cycle reset sequence
- Decimal mode does NOT clear on interrupts
- Undefined behavior for illegal opcodes

### WDC65C02 Enhancements

- All NMOS 6502 instructions plus 27 new instructions:
  - `STZ` - Store Zero
  - `BRA` - Branch Always
  - `PHX`/`PHY`/`PLX`/`PLY` - Push/Pull X and Y registers
  - `TSB`/`TRB` - Test and Set/Reset Bits
  - `BBR`/`BBS` - Branch on Bit Reset/Set
  - `RMB`/`SMB` - Reset/Set Memory Bit
  - `WAI`/`STP` - Wait for Interrupt / Stop
- 2 new addressing modes:
  - Zero Page Indirect: `($nn)`
  - Absolute Indexed Indirect: `($nnnn,X)`
- JMP indirect bug is **fixed**
- 7-cycle reset sequence
- Decimal mode **clears automatically** on interrupts
- All illegal opcodes become NOPs

## Installation

```bash
# Clone the repository
git clone https://github.com/andrewthecodertx/go-6502-emulator.git
cd go-6502-emulator

# Run tests to verify installation
go test ./...
```

## Usage

### Using the CPUs in Your Project

Add the module to your project:

```bash
go get github.com/andrewthecodertx/go-6502-emulator
```

### Basic Example - MOS6502

```go
package main

import (
    "github.com/andrewthecodertx/go-6502-emulator/pkg/mos6502"
)

// SimpleRAM implements a basic 64KB memory
type SimpleRAM struct {
    memory [0x10000]byte
}

func (r *SimpleRAM) Read(addr uint16) byte {
    return r.memory[addr]
}

func (r *SimpleRAM) Write(addr uint16, data byte) {
    r.memory[addr] = data
}

func main() {
    // Create memory
    ram := &SimpleRAM{}

    // Load a simple program at 0x8000
    // LDA #$42 - Load 0x42 into accumulator
    ram.memory[0x8000] = 0xA9 // LDA immediate opcode
    ram.memory[0x8001] = 0x42 // value
    // BRK - Break
    ram.memory[0x8002] = 0x00

    // Set reset vector to 0x8000
    ram.memory[0xFFFC] = 0x00
    ram.memory[0xFFFD] = 0x80

    // Create CPU and run
    cpu := mos6502.NewCPU(ram)
    cpu.Reset()

    // Execute a few instructions
    for i := 0; i < 10 && !cpu.Halted; i++ {
        cpu.Step()
    }

    // Check accumulator value
    println("Accumulator:", cpu.A) // Should print 0x42
}
```

### Basic Example - WDC65C02

```go
package main

import (
    "github.com/andrewthecodertx/go-6502-emulator/pkg/wdc65c02"
)

func main() {
    // Same SimpleRAM implementation as above...
    ram := &SimpleRAM{}

    // Load a program using 65C02-specific instruction
    // STZ $2000 - Store zero to address 0x2000
    ram.memory[0x8000] = 0x9C // STZ absolute opcode
    ram.memory[0x8001] = 0x00
    ram.memory[0x8002] = 0x20
    // BRK
    ram.memory[0x8003] = 0x00

    // Set reset vector
    ram.memory[0xFFFC] = 0x00
    ram.memory[0xFFFD] = 0x80

    // Create CPU and run
    cpu := wdc65c02.NewCPU(ram)
    cpu.Reset()
    cpu.Run() // Run until halted
}
```

### Implementing a Custom Bus

The `Bus` interface allows you to implement custom memory behavior:

```go
package main

import (
    "github.com/andrewthecodertx/go-6502-emulator/pkg/core"
)

// MemoryMapper with ROM and RAM regions
type MemoryMapper struct {
    ram [0x8000]byte  // RAM: 0x0000-0x7FFF
    rom [0x8000]byte  // ROM: 0x8000-0xFFFF
}

func (m *MemoryMapper) Read(addr uint16) byte {
    if addr < 0x8000 {
        return m.ram[addr]
    }
    return m.rom[addr-0x8000]
}

func (m *MemoryMapper) Write(addr uint16, data byte) {
    if addr < 0x8000 {
        m.ram[addr] = data
    }
    // ROM writes are ignored
}

// DebugBus wraps another bus and logs all accesses
type DebugBus struct {
    inner core.Bus
}

func (d *DebugBus) Read(addr uint16) byte {
    value := d.inner.Read(addr)
    println("Read:", addr, "=", value)
    return value
}

func (d *DebugBus) Write(addr uint16, data byte) {
    println("Write:", addr, "=", data)
    d.inner.Write(addr, data)
}
```

## Architecture

### Package Structure

```
go-6502-emulator/
├── pkg/
│   ├── core/             # Shared components
│   │   ├── cpu.go        # BaseCPU with common operations
│   │   ├── bus.go        # Bus interface definition
│   │   ├── variant.go    # CPU variant identification
│   │   └── addressing.go # Common addressing modes
│   ├── mos6502/          # NMOS 6502 implementation
│   │   ├── cpu.go        # NMOS 6502 CPU
│   │   ├── addressing.go # NMOS-specific addressing
│   │   └── instructions/ # Instruction implementations
│   └── wdc65c02/         # WDC 65C02 implementation
│       ├── cpu.go        # WDC 65C02 CPU
│       ├── addressing.go # 65C02-specific addressing
│       └── instructions/ # Instruction implementations
├── docs/                 # Documentation
├── CLAUDE.md             # Claude Code guidance
└── README.md
```

### CPU Architecture

Both CPU variants are built on a common `BaseCPU` structure in the `core` package:

**Registers:**

- **PC** (Program Counter): 16-bit address of next instruction
- **SP** (Stack Pointer): 8-bit offset into stack page (0x0100-0x01FF)
- **A** (Accumulator): 8-bit general purpose register
- **X, Y** (Index Registers): 8-bit index registers
- **Status**: 8-bit processor status register (flags)

**Status Flags (NV-BDIZC):**

- **N** (Negative): Set if bit 7 of result is set
- **V** (Overflow): Set on signed arithmetic overflow
- **-** (Unused): Always 1
- **B** (Break): Set by BRK instruction
- **D** (Decimal): Enables BCD arithmetic mode
- **I** (Interrupt Disable): Disables IRQ interrupts
- **Z** (Zero): Set if result is zero
- **C** (Carry): Set on unsigned overflow or borrow

**Memory Map:**

- `0x0000-0x00FF`: Zero Page (fast access)
- `0x0100-0x01FF`: Stack (grows downward from 0x01FF)
- `0x0200-0xFFF9`: General purpose memory
- `0xFFFA-0xFFFB`: NMI vector
- `0xFFFC-0xFFFD`: RESET vector
- `0xFFFE-0xFFFF`: IRQ/BRK vector

### Addressing Modes

The emulator implements all 6502 addressing modes:

1. **Implicit**: Operation on accumulator or implied register
2. **Immediate** (`#$nn`): Operand is a constant
3. **Zero Page** (`$nn`): Single-byte address in page 0
4. **Zero Page,X** (`$nn,X`): Zero page address + X register
5. **Zero Page,Y** (`$nn,Y`): Zero page address + Y register
6. **Absolute** (`$nnnn`): Full 16-bit address
7. **Absolute,X** (`$nnnn,X`): Absolute address + X register
8. **Absolute,Y** (`$nnnn,Y`): Absolute address + Y register
9. **Indirect** (`($nnnn)`): Address stored at pointer (JMP only)
10. **Indexed Indirect** (`($nn,X)`): Zero page address + X, then read pointer
11. **Indirect Indexed** (`($nn),Y`): Read zero page pointer, then add Y
12. **Relative**: Signed 8-bit offset (branch instructions)

**WDC 65C02 adds:**
13. **Zero Page Indirect** (`($nn)`): Pointer in zero page
14. **Absolute Indexed Indirect** (`($nnnn,X)`): Indexed pointer

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/mos6502
go test ./pkg/wdc65c02

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Test Organization

Tests are located alongside the code they test:

- `pkg/mos6502/cpu_test.go` - MOS6502 CPU tests
- `pkg/mos6502/hardware_test.go` - Hardware quirk tests (JMP bug, etc.)
- `pkg/wdc65c02/cpu_test.go` - WDC65C02 CPU tests

## Development

This is a library package, not a standalone executable. To use it in your project:

```bash
# Import in your Go code
go get github.com/andrewthecodertx/go-6502-emulator

# Or if developing locally
go mod tidy
```

## References

### MOS6502 (NMOS)

- [6502.org](http://www.6502.org/) - Comprehensive 6502 resources
- [Visual 6502](http://www.visual6502.org/) - Visual simulator and transistor-level analysis
- [MCS6500 Family Programming Manual](http://archive.6502.org/books/mcs6500_family_programming_manual.pdf) - Official programming manual
- [6502 Instruction Reference](http://www.6502.org/tutorials/6502opcodes.html) - Complete opcode table
- [Wikipedia: MOS Technology 6502](https://en.wikipedia.org/wiki/MOS_Technology_6502)

### WDC65C02

- [Western Design Center](https://www.westerndesigncenter.com/wdc/) - Official WDC website
- [W65C02S Datasheet](https://www.westerndesigncenter.com/wdc/documentation/w65c02s.pdf) - Official 65C02 datasheet
- [65C02 Instruction Set](https://www.westerndesigncenter.com/wdc/documentation/w65c02s_instruction_set.pdf) - Complete instruction reference
- [Wikipedia: WDC65C02](https://en.wikipedia.org/wiki/WDC_65C02)

### General Resources

- [6502 Assembly Tutorial](http://www.obelisk.me.uk/6502/) - Beginner-friendly tutorial
- [Easy 6502](https://skilldrick.github.io/easy6502/) - Interactive 6502
programming tutorial
- [NesDev Wiki](https://www.nesdev.org/wiki/CPU) - NES-specific 6502 information

## Implementation Details

### Cycle Accuracy

The emulator maintains cycle-accurate timing by:

- Tracking remaining cycles for the current instruction
- Adding extra cycles for page boundary crossings
- Correctly timing interrupt handling (7 cycles)
- Respecting variant-specific reset cycles (6 for NMOS, 7 for WDC)

### Hardware Quirks

**JMP Indirect Page Boundary Bug (MOS6502 only):**

The original 6502 has a bug where `JMP ($xxFF)` incorrectly reads the high byte
from `$xx00` instead of `$(xx+1)00`. This emulator faithfully reproduces this
behavior for MOS6502 and fixes it for WDC65C02.

```
NMOS6502:  JMP ($10FF) reads low from $10FF, high from $1000 (BUG)
WDC65C02:  JMP ($10FF) reads low from $10FF, high from $1100 (FIXED)
```

### Decimal Mode

Both processors support BCD (Binary Coded Decimal) mode for ADC and SBC
instructions. Key difference:

- **MOS6502**: Decimal flag persists through interrupts
- **WDC65C02**: Decimal flag is automatically cleared on interrupt

### Illegal Opcodes

- **MOS6502**: Undefined behavior (currently halts emulator)
- **WDC65C02**: All illegal opcodes are treated as NOPs (1 byte, 1 cycle)

## Performance

The emulator prioritizes accuracy over speed, but is still efficient enough for
real-time emulation of complete systems. Performance characteristics:

- Written in Go for memory safety and ease of development
- Minimal allocations in the main execution loop
- Bus interface allows for optimized memory implementations

## License

MIT, see [LICENSE](LICENSE).

## Contributing

PRs welcome. Please open an issue first for major changes.
