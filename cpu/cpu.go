package cpu

import (
	"github.com/edison-moreland/go6502/memory"
	"github.com/pkg/errors"
)

/*
https://en.wikipedia.org/wiki/Interrupts_in_65xx_processors
CPU is little endian

Interrupts get address from memory location:
NMI: $FFFA-FFFB
RST: $FFFC-FFFD
IRQ: #FFFE-FFFF

*/
// Register Names
const A = "A"
const X = "X"
const Y = "Y"

// Interrupt types + Locations
const NMI = "NMI"
const RST = "RST"
const IRQ = "IRQ"
const BRK = "BRK"

var interruptVectorLocations = map[string]uint16{NMI: 0xFFFA, RST: 0xFFFC, IRQ: 0xFFFE, BRK: 0xFFFE}

func stackAddress(stackPointer byte) (address uint16, err error) {
	address, err = memory.BytesToWord([2]byte{0x01, stackPointer})
	if err != nil {
		return 0, errors.Wrapf(err, "Error converting stack pointer, %#v, to real address", stackPointer)
	}

	return address, nil
}

type Go6502 struct {
	X, Y, A, SP byte
	PC          uint16
	Stat        Status
	Mem         memory.Memory

	interruptOccurred    bool
	currentInterruptType string

	CurrentInstruction Instruction

	shouldStopPCAutoIncrement bool

	shouldStopEmulation bool

	enableAddons bool
	addons       []Addon
}

func (g6 *Go6502) RegisterAddons(newAddons ...Addon) {
	if len(newAddons) > 0 {
		if g6.enableAddons == false {
			g6.enableAddons = true
		}

		g6.addons = append(g6.addons, newAddons...)

		for _, addon := range newAddons {
			addon.Register(g6)
		}
	}
}

func (g6 *Go6502) PushByteToStack(data byte) (err error) {
	// Stack pointer always refers to an address in the second page
	actualAddress, err := stackAddress(g6.SP)
	if err != nil {
		return errors.Wrapf(err, "Error pushing byte %#v to stack position %#v", data, g6.SP)
	}

	err = g6.Mem.WriteByte(actualAddress, data)
	if err != nil {
		return errors.Wrapf(err, "Error pushing byte %#v to stack position %#v", data, g6.SP)
	}
	g6.SP--

	return nil
}

func (g6 *Go6502) PopByteOffStack() (data byte, err error) {
	g6.SP++
	actualAddress, err := stackAddress(g6.SP)
	if err != nil {
		return 0, errors.Wrapf(err, "Error pulling byte off stack position %#v", g6.SP)
	}

	data, err = g6.Mem.ReadByte(actualAddress)
	if err != nil {
		return 0, errors.Wrapf(err, "Error pulling byte off stack position %#v", g6.SP)
	}

	return data, nil
}

func (g6 *Go6502) PushWordToStack(data uint16) (err error) {
	// Split word into two bytes
	bytes := memory.WordToBytes(data)

	// Push first byte
	err = g6.PushByteToStack(bytes[0])
	if err != nil {
		return errors.Wrapf(err, "Error pushing first byte of word %#v to stack", data)
	}

	// Push second byte
	err = g6.PushByteToStack(bytes[1])
	if err != nil {
		return errors.Wrapf(err, "Error pushing second byte of word %#v to stack", data)
	}

	return nil
}

func (g6 *Go6502) PopWordOffStack() (data uint16, err error) {
	// Create array to hold bytes so we can shove them into a word
	bytes := [2]byte{}

	// Bytes come off the stack in reverse order
	// Pop second byte off
	bytes[1], err = g6.PopByteOffStack()
	if err != nil {
		return 0, errors.Wrap(err, "Error popping first byte of word off stack")
	}

	// Pop first byte off
	bytes[0], err = g6.PopByteOffStack()
	if err != nil {
		return 0, errors.Wrap(err, "Error popping second byte of word off stack")
	}

	// Shove bytes into a word
	data, err = memory.BytesToWord(bytes)
	if err != nil {
		return 0, errors.Wrapf(err, "Error turning popped bytes %#v into word", bytes)
	}

	return data, nil
}

func (g6 *Go6502) GetInterruptVector(interruptType string) (uint16, error) {
	// Get interrupt vector from memory
	location := interruptVectorLocations[interruptType]

	vector, err := g6.Mem.ReadWord(location)
	if err != nil {
		return 0, errors.Wrap(err, "Error retrieving interrupt vector from mem")
	}

	return vector, nil
}

func (g6 *Go6502) HandleInterrupts() (err error) {
	/*
		Handles all 4 types of interrupts, RST IRQ NMI BRK

		if (NMI/IRQ/BRK):
			PC -> Stack
			Status -> Stack

		Set interrupt_disable
		Load vector
	*/
	var interruptType = g6.currentInterruptType

	if interruptType != RST {
		// Every interrupt but rst pushes PC and Stat onto the stack
		err = g6.nonRSTInterrupt(interruptType)
		if err != nil {
			return errors.Wrapf(err, "Error while handling non-RST interrupt, type %#v", interruptType)
		}
	}

	// Find location of interrupt handler
	interruptVector, err := g6.GetInterruptVector(interruptType)
	if err != nil {
		return errors.Wrapf(err, "Error while handling interrupt type %#v, couldn't get vector", interruptType)
	}

	// Start executing interrupt code
	g6.Stat.InterruptDisable = true
	g6.PC = interruptVector

	// Clean up
	g6.currentInterruptType = ""
	g6.interruptOccurred = false
	return nil
}

func (g6 *Go6502) nonRSTInterrupt(interruptType string) (err error) {
	// Throws PC and Status on the stack, sets bFlag if interrupt is software
	err = g6.PushWordToStack(g6.PC)
	if err != nil {
		return errors.Wrapf(err, "Error while handling interrupt type %v, couldn't push PC (%#v) to stack", interruptType, g6.PC)
	}

	// If interrupt was caused by software, bFlag needs to be set
	isSoftwareInterrupt := interruptType == BRK
	err = g6.PushByteToStack(g6.Stat.AsByte(isSoftwareInterrupt))
	if err != nil {
		return errors.Wrapf(err, "Error while handling interrupt type %v, couldn't push status to stack", interruptType)
	}

	return nil
}

func (g6 *Go6502) StopEmulation() {
	g6.shouldStopEmulation = true
}

func (g6 *Go6502) StopPCAutoIncrement() {
	g6.shouldStopPCAutoIncrement = true
}

func (g6 *Go6502) ExecuteInstruction() (err error) {
	// Find target for instruction
	addressingFunc := addressingModes[g6.CurrentInstruction.Mode]
	var targetAddress uint16
	if addressingFunc != nil {
		targetAddress, err = addressingFunc(g6)
		if err != nil {
			return errors.Wrapf(err, "Couldn't find target address for instruction (%v %v)", g6.CurrentInstruction.Mnemonic, g6.CurrentInstruction.Mode)
		}
	}

	// Execute instruction
	if instructionHandler, ok := InstructionHandlers[g6.CurrentInstruction.Mnemonic]; ok {
		if addressingFunc != nil {
			err = instructionHandler(g6, &targetAddress)
		} else {
			err = instructionHandler(g6, nil)
		}

		if err != nil {
			return errors.Wrapf(err, "Error executing instruction (%v %v)", g6.CurrentInstruction.Mnemonic, g6.CurrentInstruction.Mode)
		}
	} else {
		return errors.Errorf("Instruction (%v %v) Has not been implemented", g6.CurrentInstruction.Mnemonic, g6.CurrentInstruction.Mode)
	}

	// Increment PC
	if !g6.shouldStopPCAutoIncrement {
		g6.PC += g6.CurrentInstruction.Size
	} else {
		g6.shouldStopPCAutoIncrement = false
	}

	return nil
}

func (g6 *Go6502) StartEmulation() (err error) {
	// Trigger RESET
	g6.interruptOccurred = true
	g6.currentInterruptType = RST
	if err = g6.HandleInterrupts(); err != nil {
		return errors.Wrap(err, "Error handling RESET to start emulation")
	}

	// Start emulation
	if err = g6.emulationLoop(); err != nil {
		return err
	}

	return
}

func (g6 *Go6502) StartEmulationAtAddress(startAddress uint16) (err error) {
	// Starts emulation by setting ProgramCounter to start address
	g6.PC = startAddress

	if err = g6.emulationLoop(); err != nil {
		return err
	}

	return
}

func panicRecovery(err *error) {
	// Captures a panic and turns it into an error
	// defer at the top of a panicky function with err being a named return
	if r := recover(); r != nil {
		// Return recovered error with stacktrace
		*err = errors.Errorf("[RECOVERED PANIC]: %#v", r)
	}
}

func (g6 *Go6502) emulationLoop() (err error) {
	defer panicRecovery(&err)
	g6.shouldStopPCAutoIncrement = false

	// Turn off addons if none are registered
	if (len(g6.addons) <= 0) && g6.enableAddons {
		g6.enableAddons = false
	}

	// Emulation loop!
	for !g6.shouldStopEmulation {
		// Fetch instruction
		opcode, err := g6.Mem.ReadByte(g6.PC)
		if err != nil {
			return errors.Wrap(err, "Error retrieving instruction")
		}

		// Decode instruction
		if instruction, ok := InstructionSet[opcode]; ok {
			g6.CurrentInstruction = instruction
		} else {
			return errors.Errorf("Opcode %#v does not exist", opcode)
		}

		// Execute instruction
		if err = g6.ExecuteInstruction(); err != nil {
			return errors.Wrap(err, "Error executing instruction")
		}

		// Run AfterExecution for each addon
		if g6.enableAddons {
			for _, addon := range g6.addons {
				addon.AfterExecution()
			}

		}

		// Handle interrupts
		if g6.interruptOccurred == true {
			err = g6.HandleInterrupts()
			if err != nil {
				return errors.Wrapf(err, "Error handling interrupt after instruction %#v", opcode)
			}
		}
	}
	return
}
