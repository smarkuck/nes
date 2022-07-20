package state

import (
	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
)

const (
	DisableInterrupt = 1 << 2
	Break            = 1 << 4
	Unused           = 1 << 5

	resetVector = 0xfffc
	paramOffset = 1
)

type State struct {
	Accumulator    byte
	RegisterX      byte
	RegisterY      byte
	Status         byte
	StackPtr       byte
	ProgramCounter uint16
	nes.Bus
}

func (s *State) LoadResetProgram() {
	s.ProgramCounter = s.ReadTwoBytes(resetVector)
}

func (s *State) ReadTwoBytesParam() uint16 {
	return s.ReadTwoBytes(s.GetParamAddress())
}

func (s *State) ReadOneByteParam() byte {
	return s.Read(s.GetParamAddress())
}

func (s *State) ReadTwoBytes(addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(addr + 1)
	return byteutil.Merge(hi, lo)
}

func (s *State) ReadTwoBytesPageOverflow(
	addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(byteutil.IncrementLowByte(addr))
	return byteutil.Merge(hi, lo)
}

func (s *State) GetParamAddress() uint16 {
	return s.ProgramCounter + paramOffset
}
