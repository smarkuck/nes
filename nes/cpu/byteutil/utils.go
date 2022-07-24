package byteutil

import "fmt"

const (
	BinByte     = "%08b"
	HexByte     = "%#02x"
	TwoHexBytes = "%#04x"

	invalidBitFormat = "bits 0-7 are allowed, passed %v"
)

func IsHighEqual(addr1, addr2 uint16) bool {
	return GetHigh(addr1) == GetHigh(addr2)
}

func IncrementLow(value uint16) uint16 {
	lo := GetLow(value) + 1
	return Merge(GetHigh(value), lo)
}

func GetHigh(value uint16) byte {
	return byte(value >> 8)
}

func GetLow(value uint16) byte {
	return byte(value)
}

func Merge(hi, lo byte) uint16 {
	return uint16(hi)<<8 + uint16(lo)
}

func ToArithmeticUint16(b byte) uint16 {
	r := uint16(b)
	if IsNegative(b) {
		r |= 0xff00
	}
	return r
}

func IsNegative(b byte) bool {
	return IsLeftmostBit(b)
}

func IsLeftmostBit(b byte) bool {
	return IsBit(b, 7)
}

func IsRightmostBit(b byte) bool {
	return IsBit(b, 0)
}

func IsBit(b byte, bit uint8) bool {
	if bit > 7 {
		panic(fmt.Errorf(invalidBitFormat, bit))
	}
	return b&(1<<bit) != 0
}
