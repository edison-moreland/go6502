package cpu

import "github.com/edison-moreland/go6502/memory"

func zeroPageAddress(addr byte) (zeroPageAddr uint16, err error) {
	// Convert byte into address in page 0 of memory
	zeroPageAddr, err = memory.BytesToWord([2]byte{addr, 0x00})
	return
}

// Addressing modes
type addressingMode func(g6 *Go6502) (addr uint16, err error)

//	#	immediate	`OPC #$BB`	operand is byte BB
func immediate(g6 *Go6502) (addr uint16, err error) {
	return g6.PC + 1, nil
}

//	zpg		zeropage			`OPC $LL`	operand is zeropage address (hi-byte is zero, address = $00LL)
func zeroPage(g6 *Go6502) (addr uint16, err error) {
	// Use immediate byte as address for location in the 0th page of mem
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr, err = zeroPageAddress(immediateByte)
	if err != nil {
		return 0, err
	}

	return
}

//	zpg,X	zeropage, X-indexed	`OPC $LL,X`	operand is zeropage address; effective address is address incremented by X without carry **
func zeroPageX(g6 *Go6502) (addr uint16, err error) {
	// Use immediate byte as address for location in the 0th page of mem
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr, err = zeroPageAddress(immediateByte + g6.X)
	if err != nil {
		return 0, err
	}

	return
}

//	zpg,Y	zeropage, Y-indexed	`OPC $LL,Y`	operand is zeropage address; effective address is address incremented by Y without carry **
func zeroPageY(g6 *Go6502) (addr uint16, err error) {
	// Use immediate byte as address for location in the 0th page of mem
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr, err = zeroPageAddress(immediateByte + g6.Y)
	if err != nil {
		return 0, err
	}

	return
}

//	abs		absolute	 		 OPC $LLHH		operand is address $HHLL *
func absolute(g6 *Go6502) (addr uint16, err error) {
	// Use next two byte as address
	addr, err = g6.Mem.ReadWord(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	return
}

//	abs,X	absolute, X-indexed	 OPC $LLHH,X	operand is address; effective address is address incremented by X with carry **
func absoluteX(g6 *Go6502) (addr uint16, err error) {
	addr, err = g6.Mem.ReadWord(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr += uint16(g6.X)

	return
}

//	abs,Y	absolute, Y-indexed	 OPC $LLHH,Y	operand is address; effective address is address incremented by Y with carry **
func absoluteY(g6 *Go6502) (addr uint16, err error) {
	addr, err = g6.Mem.ReadWord(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr += uint16(g6.Y)

	return
}

//	rel	relative	`OPC $BB`	branch target is PC + signed offset BB ***
func relative(g6 *Go6502) (addr uint16, err error) {
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	targetAddress := int16(g6.PC+2) + int16(int8(immediateByte)) // Add signed offset to addr

	return uint16(targetAddress), nil
}

//	ind		indirect	 		 `OPC ($LLHH)`	operand is address; effective address is contents of word at address: C.w($HHLL)
func indirect(g6 *Go6502) (addr uint16, err error) {
	immediateWord, err := g6.Mem.ReadWord(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addr, err = g6.Mem.ReadWord(immediateWord)
	if err != nil {
		return 0, err
	}

	return
}

//	X,ind	X-indexed, indirect	 `OPC ($LL,X)`	operand is zeropage address; effective address is word in (LL + X, LL + X + 1), inc. without carry: C.w($00LL + X)
func indirectX(g6 *Go6502) (addr uint16, err error) {
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addrLoc, err := zeroPageAddress(immediateByte + g6.X)
	if err != nil {
		return 0, err
	}

	addr, err = g6.Mem.ReadWord(addrLoc)
	if err != nil {
		return 0, err
	}

	return
}

//	ind,Y	indirect, Y-indexed	 `OPC ($LL),Y`	operand is zeropage address; effective address is word in (LL, LL + 1) incremented by Y with carry: C.w($00LL) + Y
func indirectY(g6 *Go6502) (addr uint16, err error) {
	immediateByte, err := g6.Mem.ReadByte(g6.PC + 1)
	if err != nil {
		return 0, err
	}

	addrLoc, err := zeroPageAddress(immediateByte)
	if err != nil {
		return 0, err
	}

	addr, err = g6.Mem.ReadWord(addrLoc)
	if err != nil {
		return 0, err
	}

	addr += uint16(g6.Y)

	return
}

var addressingModes = map[string]addressingMode{
	"IMP":  nil, // Implied
	"ACC":  nil, // Accumulator
	"IMM":  immediate,
	"ZP":   zeroPage,
	"ZPX":  zeroPageX,
	"ZPY":  zeroPageY,
	"IND":  indirect,
	"INDX": indirectX,
	"INDY": indirectY,
	"ABS":  absolute,
	"ABSX": absoluteX,
	"ABSY": absoluteY,
	"REL":  relative,
}
