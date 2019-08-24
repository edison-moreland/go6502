package memory

import (
	"encoding/binary"
	"github.com/pkg/errors"
	"os"
)

func WordToBytes(word uint16) (bytes [2]byte) {
	// takes word and turns it into an array of two bytes in little endian order
	bytes = *new([2]byte)
	binary.BigEndian.PutUint16(bytes[:], word)

	// Reverse byte order
	bytes[0], bytes[1] = bytes[1], bytes[0]

	return bytes
}

func BytesToWord(bytes [2]byte) (word uint16, err error) {
	// takes bytes in little endian order and turns it into a word
	//defer panicRecovery(&err)
	word = binary.LittleEndian.Uint16(bytes[:])
	return word, nil
}

type Memory struct {
	// +1 so the full range of 16bit numbers can be used as an address
	// Just 0xFFFF would mean Memory.mem[0xFFFF] causes an out of bounds
	Mem [0xFFFF + 1]byte
}

func (m *Memory) LoadMem(path string, startAddress uint16, endAddress uint16) (err error) {
	// Load file at path into 6502 memory, WARNING: OVERWRITES MEMORY
	file, err := os.Open(path)
	defer file.Close() // make sure file get's closed

	if err != nil {
		err = errors.WithStack(err)
		return errors.Wrapf(err, "Error opening file: %v", path)
	}

	// Read file contents to mem
	_, err = file.Read(m.Mem[startAddress:endAddress])
	if err != nil {
		err = errors.WithStack(err)
		return errors.Wrapf(err, "Error reading file: %v", path)
	}

	return nil
}

func (m *Memory) ReadWord(loc uint16) (word uint16, err error) {
	//defer panicRecovery(&err)

	// Trying to get a slice of Mem (g6.Mem[loc:loc+1]) would sometimes only return one number
	// So instead grab them both separately and throw them into a slice
	rawWord := [2]byte{m.Mem[loc], m.Mem[loc+1]}
	word, err = BytesToWord(rawWord)
	if err != nil {
		return 0, errors.Wrapf(err, "Error converting raw bytes to word: %v", rawWord)
	}

	return word, nil
}

func (m *Memory) WriteWord(loc uint16, word uint16) (err error) {
	//defer panicRecovery(&err)
	// TODO: Undo any writes after a panic

	// Separate word into two bytes
	rawWord := WordToBytes(word)

	// Write both bytes to mem
	m.Mem[loc] = rawWord[0]
	m.Mem[loc+1] = rawWord[1]

	return nil
}

func (m *Memory) ReadByte(loc uint16) (memByte byte, err error) {
	//defer panicRecovery(&err)
	memByte = m.Mem[loc]
	return memByte, nil
}

func (m *Memory) WriteByte(loc uint16, memByte byte) (err error) {
	//defer panicRecovery(&err)
	m.Mem[loc] = memByte
	return nil
}
