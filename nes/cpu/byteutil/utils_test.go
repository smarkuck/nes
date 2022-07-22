package byteutil_test

import (
	"fmt"

	. "github.com/smarkuck/nes/nes/cpu/byteutil"
	. "github.com/smarkuck/unittest"
)

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

func Test_IsRightmost(t *T) {
	ExpectFalse(t, IsRightmost(0b00000000))
	ExpectFalse(t, IsRightmost(0b11111110))
	ExpectTrue(t, IsRightmost(0b00000001))
	ExpectTrue(t, IsRightmost(0b11111111))
}

func Test_IsLeftmost(t *T) {
	ExpectFalse(t, IsLeftmost(0b00000000))
	ExpectFalse(t, IsLeftmost(0b01111111))
	ExpectTrue(t, IsLeftmost(0b10000000))
	ExpectTrue(t, IsLeftmost(0b11111111))
}

func Test_IsNegative(t *T) {
	ExpectFalse(t, IsNegative(0x00))
	ExpectFalse(t, IsNegative(0x7f)) // 127
	ExpectTrue(t, IsNegative(0x80))  // -128
	ExpectTrue(t, IsNegative(0xff))  // -1
}

func Test_ToArithmeticUint16(t *T) {
	ExpectEq(t, ToArithmeticUint16(0x00), 0x0000)
	ExpectEq(t, ToArithmeticUint16(0x7f), 0x007f)
	ExpectEq(t, ToArithmeticUint16(0x80), 0xff80)
	ExpectEq(t, ToArithmeticUint16(0xff), 0xffff)
}
