package byteutil

const (
	BinByte     = "%08b"
	HexByte     = "%#02x"
	TwoHexBytes = "%#04x"
)

func IsSameHighByte(addr1, addr2 uint16) bool {
	return GetHigh(addr1) == GetHigh(addr2)
}

func IncrementLowByte(value uint16) uint16 {
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

func ToArithmeticUint16(shift byte) uint16 {
	s := uint16(shift)
	if IsNegative(shift) {
		s |= 0xff00
	}
	return s
}

func IsNegative(b byte) bool {
	return b&0x80 != 0
}
