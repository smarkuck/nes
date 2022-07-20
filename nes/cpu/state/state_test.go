package state_test

import (
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	. "github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	resetVector    = 0xfffc
	programCounter = 0xcafe
	paramOffset    = 1
	value          = 173
	value16High    = 0x2f
	value16Low     = 0x9c
	value16        = 0x2f9c
)

func Test_GetParamAddress(t *T) {
	s := State{ProgramCounter: programCounter}

	ExpectEqf(t,
		s.GetParamAddress(), programCounter+paramOffset,
		byteutil.TwoHexBytes)
}

func Test_ReadTwoBytes(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xcb00: value16High,
	}}

	ExpectEqf(t, s.ReadTwoBytes(0xcaff), value16,
		byteutil.TwoHexBytes)
}

func Test_ReadTwoBytesPageOverflow(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xca00: value16High,
	}}

	ExpectEqf(t,
		s.ReadTwoBytesPageOverflow(0xcaff), value16,
		byteutil.TwoHexBytes)
}

func Test_ReadOneByteParam(t *T) {
	s := State{
		ProgramCounter: programCounter,
		Bus: TestBus{
			programCounter + paramOffset: value,
		},
	}

	ExpectEq(t, s.ReadOneByteParam(), value)
}

func Test_ReadTwoBytesParam(t *T) {
	s := State{
		ProgramCounter: programCounter,
		Bus: TestBus{
			programCounter + paramOffset:     value16Low,
			programCounter + paramOffset + 1: value16High,
		},
	}

	ExpectEqf(t, s.ReadTwoBytesParam(), value16,
		byteutil.TwoHexBytes)
}

func Test_LoadResetProgram(t *T) {
	s := State{Bus: TestBus{
		resetVector:     value16Low,
		resetVector + 1: value16High,
	}}

	s.LoadResetProgram()

	ExpectEqf(t, s.ProgramCounter, value16,
		byteutil.TwoHexBytes)
}
