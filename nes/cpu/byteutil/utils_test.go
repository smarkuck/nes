package byteutil_test

import (
	"fmt"

	. "github.com/smarkuck/nes/nes/cpu/byteutil"
	. "github.com/smarkuck/unittest"
)

const invalidBitFormat = "bits 0-7 are allowed, passed %v"

func getInvalidBitText(value uint8) string {
	return fmt.Sprintf(invalidBitFormat, value)
}

func Test_Formats(t *T) {
	ExpectEq(t, fmt.Sprintf(BinByte, 11), "00001011")
	ExpectEq(t, fmt.Sprintf(HexByte, 11), "0x0b")
	ExpectEq(t, fmt.Sprintf(TwoHexBytes, 11), "0x000b")
}

func Test_GetLow(t *T) {
	ExpectEqf(t, GetLow(0x8cfa), 0xfa, HexByte)
}

func Test_GetHigh(t *T) {
	ExpectEqf(t, GetHigh(0x8cfa), 0x8c, HexByte)
}

func Test_Merge(t *T) {
	ExpectEqf(t, Merge(0x8c, 0xfa), 0x8cfa, TwoHexBytes)
}

func Test_IsHighEqual(t *T) {
	ExpectTrue(t, IsHighEqual(0x7c8a, 0x7c02))
	ExpectFalse(t, IsHighEqual(0x7d02, 0x7c02))
}

func Test_IncrementLow(t *T) {
	ExpectEqf(t, IncrementLow(0x8cfa), 0x8cfb,
		TwoHexBytes)
	ExpectEqf(t, IncrementLow(0x8cff), 0x8c00,
		TwoHexBytes)
}

func Test_IsBit_PanicOnBitGreaterThan7(t *T) {
	value := uint8(8)
	text := getInvalidBitText(value)

	defer ExpectPanicErrEq(t, text)

	IsBit(0, value)
}

func Test_IsBit(t *T) {
	ExpectTrue(t, IsBit(0b00000100, 2))
	ExpectFalse(t, IsBit(0b11111011, 2))
	ExpectTrue(t, IsBit(0b00100000, 5))
	ExpectFalse(t, IsBit(0b11011111, 5))
}

func Test_IsRightmostBit(t *T) {
	ExpectFalse(t, IsRightmostBit(0b00000000))
	ExpectFalse(t, IsRightmostBit(0b11111110))
	ExpectTrue(t, IsRightmostBit(0b00000001))
	ExpectTrue(t, IsRightmostBit(0b11111111))
}

func Test_IsLeftmostBit(t *T) {
	ExpectFalse(t, IsLeftmostBit(0b00000000))
	ExpectFalse(t, IsLeftmostBit(0b01111111))
	ExpectTrue(t, IsLeftmostBit(0b10000000))
	ExpectTrue(t, IsLeftmostBit(0b11111111))
}

func Test_IsNegative(t *T) {
	ExpectFalse(t, IsNegative(0x00))
	ExpectFalse(t, IsNegative(0x7f))
	ExpectTrue(t, IsNegative(0x80))
	ExpectTrue(t, IsNegative(0xff))
}

func Test_ToArithmeticUint16(t *T) {
	ExpectEq(t, ToArithmeticUint16(0x00), 0x0000)
	ExpectEq(t, ToArithmeticUint16(0x7f), 0x007f)
	ExpectEq(t, ToArithmeticUint16(0x80), 0xff80)
	ExpectEq(t, ToArithmeticUint16(0xff), 0xffff)
}
