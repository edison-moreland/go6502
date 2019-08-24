package testingHelp

import "math/rand"

func RandomUint16() uint16 {
	return uint16(rand.Intn(0xFFFF - 1))
}

func RandomUint8() uint8 {
	return uint8(rand.Intn(0xFF))
}

func RandomWords(len int) (words []uint16) {
	for i := 0; i <= len; i++ {
		words = append(words, RandomUint16())
	}
	return words
}
