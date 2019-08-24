package cpu

import (
	"fmt"
	"time"
)

type Addon interface {
	Register(g6 *Go6502)
	AfterExecution()
}

type BaseAddon struct {
	G6 *Go6502
}

func (ba *BaseAddon) Register(g6 *Go6502) {
	// Remember CPU for later
	ba.G6 = g6
}

func (ba BaseAddon) AfterExecution() {
	return
}

// Example addon, prints debug information after every instruction executes
type DebugAddon struct {
	BaseAddon
	SlowDown     time.Duration
	Step, ShowZP bool
}

func (da DebugAddon) AfterExecution() {
	fmt.Print("\n")
	fmt.Printf("Result of instruction %#v(%v, %v) \n", da.G6.CurrentInstruction.Opcode, da.G6.CurrentInstruction.Mnemonic, da.G6.CurrentInstruction.Mode)
	fmt.Printf("Neg: %#v, Ovr: %#v, Dec: %#v, Int: %#v, Zer: %#v, Car: %#v  \n", da.G6.Stat.Negative, da.G6.Stat.Overflow, da.G6.Stat.Decimal, da.G6.Stat.InterruptDisable, da.G6.Stat.Zero, da.G6.Stat.Carry)
	fmt.Printf("PC: %#v, SP: %#v, A: %#v, X: %#v, Y: %#v \n", da.G6.PC, da.G6.SP, da.G6.A, da.G6.X, da.G6.Y)

	if da.ShowZP {
		fmt.Print("Zeropage:")
		fmt.Printf("0x0000: %v", da.G6.Mem.Mem[0x0000:0x000F])
		fmt.Printf("0x0010: %v", da.G6.Mem.Mem[0x0010:0x001F])
		fmt.Printf("0x0020: %v", da.G6.Mem.Mem[0x0020:0x002F])
		fmt.Printf("0x0030: %v", da.G6.Mem.Mem[0x0030:0x003F])
		fmt.Printf("0x0040: %v", da.G6.Mem.Mem[0x0040:0x004F])
		fmt.Printf("0x0050: %v", da.G6.Mem.Mem[0x0050:0x005F])
		fmt.Printf("0x0060: %v", da.G6.Mem.Mem[0x0060:0x006F])
		fmt.Printf("0x0070: %v", da.G6.Mem.Mem[0x0070:0x007F])
		fmt.Printf("0x0080: %v", da.G6.Mem.Mem[0x0080:0x008F])
		fmt.Printf("0x0090: %v", da.G6.Mem.Mem[0x0090:0x009F])
		fmt.Printf("0x00A0: %v", da.G6.Mem.Mem[0x00A0:0x00AF])
		fmt.Printf("0x00B0: %v", da.G6.Mem.Mem[0x00B0:0x00BF])
		fmt.Printf("0x00C0: %v", da.G6.Mem.Mem[0x00C0:0x00CF])
		fmt.Printf("0x00D0: %v", da.G6.Mem.Mem[0x00D0:0x00DF])
		fmt.Printf("0x00E0: %v", da.G6.Mem.Mem[0x00E0:0x00EF])
		fmt.Printf("0x00F0: %v", da.G6.Mem.Mem[0x00F0:0x00FF])
	}

	if da.Step {
		// Wait for input to allow stepping through program
		_, _ = fmt.Scanln()

	} else {
		// Artificially slow down execution to make watching output easier
		if da.SlowDown != 0 {
			time.Sleep(da.SlowDown)
		}
	}
}
