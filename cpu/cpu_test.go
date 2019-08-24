package cpu

import (
	"github.com/edison-moreland/go6502/testingHelp"
	"testing"
)

// InterruptDisable Vector test
type testInterruptVector struct {
	loc           uint16
	interruptType string
}

var interruptTests = []testInterruptVector{
	{0xFFFA, "NMI"},
	{0xFFFC, "RST"},
	{0xFFFE, "IRQ"},
}

func TestGo6502_GetInterruptVector(t *testing.T) {
	for _, testData := range interruptTests {
		// Setup
		cpu := new(Go6502)
		cpu.Mem.WriteWord(testData.loc, 0x9559)

		// Test
		vector, err := cpu.GetInterruptVector(testData.interruptType)
		testingHelp.NotNil(t, err)
		testingHelp.Equals(t, uint16(0x9559), vector)
	}
}

func randomBytes(len int) (bytes []byte) {
	for i := 0; i <= len; i++ {
		bytes = append(bytes, testingHelp.RandomUint8())
	}
	return bytes
}

func TestGo6502_PushPopByteToStack(t *testing.T) {
	cpu := new(Go6502)
	cpu.SP = 0xFF // Initialize stack pointer to top of stack

	testSize := 240
	testBytes := randomBytes(testSize)

	// Push all testBytes onto the stack
	for _, randByte := range testBytes {
		err := cpu.PushByteToStack(randByte)
		testingHelp.NotNil(t, err)
	}

	// Pop testBytes off the stack
	var actBytes []byte
	for i := 0; i <= testSize; i++ {
		actByte, err := cpu.PopByteOffStack()
		testingHelp.NotNil(t, err)

		// Add actByte to front of actBytes
		actBytes = append([]byte{actByte}, actBytes...)
	}

	testingHelp.Equals(t, testBytes, actBytes)

}

func TestGo6502_PushPopWordToStack(t *testing.T) {
	cpu := new(Go6502)
	cpu.SP = 0xFF // Initialize stack pointer to top of stack

	testSize := 120
	testWords := testingHelp.RandomWords(testSize)

	// Push all testBytes onto the stack
	for _, randWord := range testWords {
		err := cpu.PushWordToStack(randWord)
		testingHelp.NotNil(t, err)
	}

	// Pop testBytes off the stack
	var actWords []uint16
	for i := 0; i <= testSize; i++ {
		actWord, err := cpu.PopWordOffStack()
		testingHelp.NotNil(t, err)

		// Add actWords to front of actWords
		actWords = append([]uint16{actWord}, actWords...)
	}

	testingHelp.Equals(t, testWords, actWords)

}
