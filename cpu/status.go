package cpu

/*
The B flag
----------
No actual "B" flag exists inside the 6502's processor status register. The B
flag only exists in the status flag byte pushed to the stack. Naturally,
when the status are restored (via PLP or RTI), the B bit is discarded.

Depending on the means, the B status flag will be pushed to the stack as
either 0 or 1.

software instructions BRK & PHP will push the B flag as being 1.
hardware interrupts IRQ & NMI will push the B flag as being 0.
*/

type Status struct {
	/*
		Bit 7: Negative
		Bit 6: Overflow
		Bit 5: Always set
		Bit 4: Break Flag
		Bit 3: Decimal mode
		Bit 2: InterruptDisable disable
		Bit 1: Zero
		Bit 0: Carry
	*/
	Negative, Overflow, Decimal, InterruptDisable, Zero, Carry bool
}

func (f *Status) AsByte(bFlag bool) byte {
	// Set bFlag true if interrupt caused by software, false if caused by hardware
	// Export status to a byte for pushing onto stack
	// Put status in order in a slice, then bitwise them into a byte
	status := [8]bool{f.Negative, f.Overflow, true, bFlag, f.Decimal, f.InterruptDisable, f.Zero, f.Carry}
	statusByte := byte(0)

	for i := uint(0); i < 8; i++ {
		flag := status[i]
		if flag {
			// Set ith bit
			statusByte |= 0x80 >> i
		}
	}
	return statusByte
}

func (f *Status) FromByte(statusByte byte) {
	// Load status from a byte for pulling from stack
	status := [8]*bool{&f.Negative, &f.Overflow, nil, nil, &f.Decimal, &f.InterruptDisable, &f.Zero, &f.Carry}

	for i := uint(0); i < 8; i++ {
		if status[i] == nil {
			// Skip unused bit
			continue
		}

		// Set ith flag to ith bit
		bit := (statusByte >> (7 - i)) & 1
		*status[i] = bool(bit == 1)
	}
}
