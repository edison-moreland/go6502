package vic2

import (
	"fmt"
	"github.com/edison-moreland/go6502/cpu"
)

// Vic-2 Register locations
const ControlRegister1 = 0xD011
const ControlRegister2 = 0xD016
const RasterCounter = 0xD012

// Vic2Addon emulates the behevior of the Video interface chip used in the Commodore ^$
type Addon struct {
	cpu.BaseAddon
}

func (v2 *Addon) AfterExecution() {
	control1, _ := v2.G6.Mem.ReadByte(ControlRegister1)
	control2, _ := v2.G6.Mem.ReadByte(ControlRegister2)
	rasterCounter, _ := v2.G6.Mem.ReadByte(RasterCounter)
	fmt.Printf("Vic20 - CR1: (%#v)  CR2: (%#v)  RC: (%v) ", control1, control2, rasterCounter)
}
