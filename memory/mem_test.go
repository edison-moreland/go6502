package memory

import (
	"github.com/edison-moreland/go6502/testingHelp"
	"testing"
)

func addressesToTest() []uint16 {
	testAddresses := []uint16{
		0x0000, // Make sure addresses on both ends of memory get tested
		0xFFFE, // The very end has been troublesome before
	}

	testAddresses = append(testingHelp.RandomWords(1000), testAddresses...)

	return testAddresses
}

func TestWordToBytes(t *testing.T) {
	// TODO: Randomize test
	testWord := uint16(0x9559)
	expBytes := [2]byte{0x59, 0x95}

	testingHelp.Equals(t, expBytes, WordToBytes(testWord))
}

func TestBytesToWord(t *testing.T) {
	// TODO: Randomize test
	testBytes := [2]byte{0x59, 0x95}
	expWord := uint16(0x9559)

	actWord, err := BytesToWord(testBytes)
	testingHelp.NotNil(t, err)
	testingHelp.Equals(t, expWord, actWord)
}

func TestMemory_ReadWriteWord(t *testing.T) {
	// Since ReadWord() and WriteWord() are so tied together they get tested together

	memory := new(Memory)
	for _, testAddress := range addressesToTest() {
		// Write random word to memory and read it back then compare
		testWord := testingHelp.RandomUint16()

		err := memory.WriteWord(testAddress, testWord)
		testingHelp.NotNil(t, err)

		word, err := memory.ReadWord(testAddress)
		testingHelp.NotNil(t, err)
		testingHelp.Equals(t, testWord, word)
	}
}

func TestMemory_ReadWriteByte(t *testing.T) {
	// Same as test above

	memory := new(Memory)

	// Another address is added on to make sure the end of memory gets hit
	for _, testAddress := range append(addressesToTest(), 0xFFFF) {
		// Write random word to memory and read it back then compare
		testByte := testingHelp.RandomUint8()

		err := memory.WriteByte(testAddress, testByte)
		testingHelp.NotNil(t, err)

		actByte, err := memory.ReadByte(testAddress)
		testingHelp.NotNil(t, err)
		testingHelp.Equals(t, testByte, actByte)
	}
}
