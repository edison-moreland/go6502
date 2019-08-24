package cpu

import (
	"github.com/edison-moreland/go6502/testingHelp"
	"testing"
)

// Test data, gives matching status and byte
type statusTestData struct {
	status Status
	asByte byte
}

var statusTests = []statusTestData{
	{Status{}, 0x20},                       // 0b0010 0000
	{Status{Negative: true}, 0xA0},         // 0b1010 0000
	{Status{Overflow: true}, 0x60},         // 0b0110 0000
	{Status{Decimal: true}, 0x28},          // 0b0010 1000
	{Status{InterruptDisable: true}, 0x24}, // 0b0010 0100
	{Status{Zero: true}, 0x22},             // 0b0010 0010
	{Status{Carry: true}, 0x21},            // 0b0010 0001
}

func TestStatus_AsByte(t *testing.T) {
	// Test Status.asByte() returns the proper byte
	for _, testData := range statusTests {
		statusByte := testData.status.AsByte(false)
		testingHelp.Equals(t, statusByte, testData.asByte)
	}
}

func TestStatus_AsByte_bFlag(t *testing.T) {
	// Test Status.AsByte() will set bit 4 when bFlag is set
	status := Status{}
	testingHelp.Equals(t, status.AsByte(false), byte(0x20)) // Empty status without bflag == 0x20
	testingHelp.Equals(t, status.AsByte(true), byte(0x30))  // Empty status with bFlag == 0x30
}

func TestStatus_FromByte(t *testing.T) {
	// Test setting status with Status.FromByte() sets the correct flags
	for _, testData := range statusTests {
		statusFromByte := Status{}
		statusFromByte.FromByte(testData.asByte)
		testingHelp.Equals(t, statusFromByte, testData.status)
	}
}
