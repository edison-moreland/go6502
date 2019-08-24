package cpu

import "github.com/pkg/errors"

// Flag calculations //
func setZeroNegativeFlags(g6 *Go6502, val uint8) {
	// Common combo of flags
	g6.Stat.Zero = val == 0          // Set zero flag
	g6.Stat.Negative = int8(val) > 0 // Set negative flag
}

func setCarryFlag(g6 *Go6502, val uint16) {
	g6.Stat.Carry = (val & 0xFF00) != 0
}

func setOverflowFlag(g6 *Go6502, val uint8, originalVal uint8) {
	g6.Stat.Overflow = (originalVal & 0x80) != (val & 0x80)
}

// Instruction Handlers
type InstructionHandler func(g6 *Go6502, targetAddress *uint16) (err error)

var InstructionHandlers = map[string]InstructionHandler{
	// ADC, Add memory to accumulator with carry
	"ADC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Add it all up, cast to uint16 to calculate carry
		total := uint16(g6.A) + uint16(val)
		if g6.Stat.Carry {
			total++
		}

		// Calculate flags
		setCarryFlag(g6, total)
		setZeroNegativeFlags(g6, uint8(total))
		setOverflowFlag(g6, uint8(total), g6.A)

		// Put back in accumulator
		g6.A = uint8(total)

		return
	},

	// AND, AND memory with accumulator
	"AND": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// And with accumulator
		g6.A &= val

		setZeroNegativeFlags(g6, g6.A)
		return
	},

	// ASL, Shift left one bit
	"ASL": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// This instruction can act on either memory or a register
		var val uint8
		if g6.CurrentInstruction.Mode == "ACC" {
			val = g6.A
		} else {
			// Read byte from target address
			val, err = g6.Mem.ReadByte(*targetAddress)
			if err != nil {
				return err
			}
		}

		// Shift left by one bit, convert to uint16 to calculate carry
		shiftedVal := uint16(val) << 1
		val = uint8(shiftedVal)

		if g6.CurrentInstruction.Mode == "ACC" {
			g6.A = val
		} else {
			// Write byte to target address
			err = g6.Mem.WriteByte(*targetAddress, val)
			if err != nil {
				return err
			}
		}

		// Calculate flags
		setZeroNegativeFlags(g6, val)
		setCarryFlag(g6, shiftedVal)

		return
	},

	// Branch InstructionSet //
	// BCC, Branch on carry clear
	"BCC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if !g6.Stat.Carry {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BCS, Branch on carry set
	"BCS": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if g6.Stat.Carry {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BEQ, Branch on result zero
	"BEQ": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if g6.Stat.Zero {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BMI, Branch on result minus
	"BMI": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if g6.Stat.Negative {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BNE, Branch on result not zero
	"BNE": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if !g6.Stat.Zero {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BPL, Branch on result plus
	"BPL": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if !g6.Stat.Negative {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BVC, Branch on overflow clear
	"BVC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if !g6.Stat.Overflow {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BVS, Branch on overflow set
	"BVS": func(g6 *Go6502, targetAddress *uint16) (err error) {
		if g6.Stat.Overflow {
			g6.PC = *targetAddress
			g6.shouldStopPCAutoIncrement = true
		}
		return nil
	},

	// BIT

	// BRK
	"BRK": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.interruptOccurred = true
		g6.currentInterruptType = BRK
		return nil
	},

	// Clear InstructionSet //
	// CLC, Clear Carry Flag
	"CLC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.Carry = false
		return nil
	},

	// CLD, Clear Decimal Mode
	"CLD": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.Decimal = false
		return nil
	},

	// CLI, Clear Interrupt Disable
	"CLI": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.InterruptDisable = false
		return nil
	},

	// CLV, Clear Overflow Flag
	"CLV": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.Overflow = false
		return nil
	},

	// NOP
	"NOP": func(g6 *Go6502, targetAddress *uint16) (err error) {
		return nil
	},

	// PHA

	// PLA

	// PHP

	// PLP

	// Return Instructions //
	// RTI, Return from interrupt
	"RTI": func(g6 *Go6502, targetAddress *uint16) (err error) {
		statusRegister, err := g6.PopByteOffStack()
		if err != nil {
			return errors.Wrap(err, "Couldn't retrieve status register from stack")
		}

		g6.Stat.FromByte(statusRegister)

		returnAddress, err := g6.PopWordOffStack()
		if err != nil {
			return errors.Wrap(err, "Couldn't retrieve return address from stack")
		}

		g6.PC = returnAddress
		g6.shouldStopPCAutoIncrement = true
		return
	},

	// RTS, Return from subroutine
	"RTS": func(g6 *Go6502, targetAddress *uint16) (err error) {
		returnAddress, err := g6.PopWordOffStack()
		if err != nil {
			return errors.Wrap(err, "Couldn't retrieve return address from stack")
		}

		g6.PC = returnAddress
		g6.shouldStopPCAutoIncrement = true
		return
	},

	// Set Flag Instructions //
	// SEC, Set Carry
	"SEC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.Carry = true
		return nil
	},

	// SED, Set Decimal
	"SED": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.Decimal = true
		return nil
	},

	// SEI, Set Interrupt Disable
	"SEI": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Stat.InterruptDisable = true
		return nil
	},

	// Transfer InstructionSet //
	// TAX, Transfer Accumulator to Index X
	"TAX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.X = g6.A

		setZeroNegativeFlags(g6, g6.A)
		return nil
	},

	// TAY, Transfer Accumulator to Index Y
	"TAY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Y = g6.A

		setZeroNegativeFlags(g6, g6.A)
		return nil
	},

	// TSX, Transfer Stack Pointer to Index x
	"TSX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.X = g6.SP

		setZeroNegativeFlags(g6, g6.SP)
		return nil
	},

	// TXA, Transfer Index X to Accumulator
	"TXA": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.A = g6.X

		setZeroNegativeFlags(g6, g6.X)
		return nil
	},

	// TXS, Transfer Index X to Stack Register
	"TXS": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.SP = g6.X

		setZeroNegativeFlags(g6, g6.X)
		return nil
	},

	// TYA, Transfer Index Y to Accumulator
	"TYA": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.A = g6.Y

		setZeroNegativeFlags(g6, g6.Y)
		return nil
	},

	// Compare InstructionSet //
	// CMP, Compare memory with accumulator
	"CMP": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Cast to uint16 to calculate carry
		comp := uint16(g6.A) - uint16(val)

		setZeroNegativeFlags(g6, uint8(comp))
		setCarryFlag(g6, comp)
		return nil
	},

	// CPX Compare memory and index x
	"CPX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Cast to uint16 to calculate carry
		comp := uint16(g6.X) - uint16(val)

		setZeroNegativeFlags(g6, uint8(comp))
		setCarryFlag(g6, comp)
		return nil
	},

	// CPY, Compare memory and index y
	"CPY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Cast to uint16 to calculate carry
		comp := uint16(g6.Y) - uint16(val)

		setZeroNegativeFlags(g6, uint8(comp))
		setCarryFlag(g6, comp)
		return nil
	},

	// Decrement InstructionSet //
	// DEC, Decrement memory by one
	"DEC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Decrement value
		val--

		// Write back to target address
		err = g6.Mem.WriteByte(*targetAddress, val)
		if err != nil {
			return err
		}

		setZeroNegativeFlags(g6, val) // Set status register flags
		return
	},

	// DEX, Decrement Index X by one
	"DEX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.X--

		setZeroNegativeFlags(g6, g6.X)
		return nil
	},

	// DEY, Decrement Index Y by one
	"DEY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Y--

		setZeroNegativeFlags(g6, g6.Y)
		return nil
	},

	// Increment Instructions //
	// INC
	"INC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Increment value
		val++

		// Write back to target address
		err = g6.Mem.WriteByte(*targetAddress, val)
		if err != nil {
			return err
		}

		setZeroNegativeFlags(g6, val) // Set status register flags
		return
	},

	// INX
	"INX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.X++

		setZeroNegativeFlags(g6, g6.X)
		return nil
	},

	// INY
	"INY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.Y++

		setZeroNegativeFlags(g6, g6.Y)
		return nil
	},

	// EOR, Exclusive-OR memory with accumulator
	"EOR": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		g6.A ^= val

		setZeroNegativeFlags(g6, g6.A)
		return nil
	},

	// Jump InstructionSet //
	// JMP, Jump to new location
	"JMP": func(g6 *Go6502, targetAddress *uint16) (err error) {
		g6.PC = *targetAddress
		g6.shouldStopPCAutoIncrement = true
		return
	},

	// JSR, Jump to new location saving return address
	"JSR": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Save address of next instruction for return
		if err = g6.PushWordToStack(g6.PC + g6.CurrentInstruction.Size); err != nil {
			return err
		}

		g6.PC = *targetAddress
		g6.shouldStopPCAutoIncrement = true
		return
	},

	// Load InstructionSet //
	// LDA, Load A with Memory
	"LDA": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Set A register
		g6.A = val

		setZeroNegativeFlags(g6, val) // Set status register flags
		return
	},

	// LDX, Load X with Memory
	"LDX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		g6.X = val

		setZeroNegativeFlags(g6, val) // Set status register flags
		return
	},

	// LDY, Load Y with Memory
	"LDY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		g6.Y = val

		setZeroNegativeFlags(g6, val) // Set status register flags
		return
	},

	// ORA, OR memory with accumulator
	"ORA": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// OR with accumulator
		g6.A |= val

		setZeroNegativeFlags(g6, g6.A)
		return
	},

	// Rotate Instructions
	// ROL, Rotate one bit left (Memory or Accumulator)
	"ROL": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// This instruction can act on either memory or a register
		var val uint8
		if g6.CurrentInstruction.Mode == "ACC" {
			val = g6.A
		} else {
			// Read byte from target address
			val, err = g6.Mem.ReadByte(*targetAddress)
			if err != nil {
				return err
			}
		}

		// TODO: Rotate, not shift
		// Shift left by one bit, convert to uint16 to calculate carry
		shiftedVal := uint16(val) << 1
		val = uint8(shiftedVal)

		if g6.CurrentInstruction.Mode == "ACC" {
			g6.A = val
		} else {
			// Write byte to target address
			err = g6.Mem.WriteByte(*targetAddress, val)
			if err != nil {
				return err
			}
		}

		// Calculate flags
		setZeroNegativeFlags(g6, val)
		setCarryFlag(g6, shiftedVal)

		return
	},

	// ROR, rotate one bit right (memory or accumulator)
	"ROR": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// This instruction can act on either memory or a register
		var val uint8
		if g6.CurrentInstruction.Mode == "ACC" {
			val = g6.A
		} else {
			// Read byte from target address
			val, err = g6.Mem.ReadByte(*targetAddress)
			if err != nil {
				return err
			}
		}

		// TODO: Rotate, not shift
		// Shift left by one bit, convert to uint16 to calculate carry
		shiftedVal := uint16(val) >> 1
		val = uint8(shiftedVal)

		if g6.CurrentInstruction.Mode == "ACC" {
			g6.A = val
		} else {
			// Write byte to target address
			err = g6.Mem.WriteByte(*targetAddress, val)
			if err != nil {
				return err
			}
		}

		// Calculate flags
		setZeroNegativeFlags(g6, val)
		setCarryFlag(g6, shiftedVal)

		return
	},

	// SBC, Subtract accumulator with memory
	"SBC": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Read byte from target address
		val, err := g6.Mem.ReadByte(*targetAddress)
		if err != nil {
			return err
		}

		// Add it all up, cast to uint16 to calculate carry
		total := uint16(g6.A) - uint16(val)
		if g6.Stat.Carry {
			total--
		}

		// Calculate flags
		setCarryFlag(g6, total)
		setZeroNegativeFlags(g6, uint8(total))
		setOverflowFlag(g6, uint8(total), g6.A)

		// Put back in accumulator
		g6.A = uint8(total)

		return
	},

	// Store InstructionSet //
	// STA, Store A in Memory
	"STA": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Write byte to target address
		err = g6.Mem.WriteByte(*targetAddress, g6.A)
		if err != nil {
			return err
		}
		return
	},

	// STX, Store X in Memory
	"STX": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Write byte to target address
		err = g6.Mem.WriteByte(*targetAddress, g6.X)
		if err != nil {
			return err
		}
		return
	},

	// STY, Store Y in Memory
	"STY": func(g6 *Go6502, targetAddress *uint16) (err error) {
		// Write byte to target address
		err = g6.Mem.WriteByte(*targetAddress, g6.Y)
		if err != nil {
			return err
		}
		return
	},
}
