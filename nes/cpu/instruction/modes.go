package instruction

import (
	"github.com/smarkuck/nes/nes/cpu"
)

type instruction = cpu.Instruction

type impliedCmd = func(*cpu.State)
type addressCmd = func(_ *cpu.State, addr uint16)
type relativeCmd = func(status byte) bool

type impliedMode struct {
	cmd    impliedCmd
	cycles uint8
}

func NewImplied(c impliedCmd, cycles uint8) instruction {
	return &impliedMode{c, cycles}
}

func NewAccumulative(c impliedCmd, cycles uint8) instruction {
	return &impliedMode{c, cycles}
}

func (i *impliedMode) Execute(s *cpu.State) {
	s.ProgramCounter++
	i.cmd(s)
}

func (i *impliedMode) GetCycles() uint8 {
	return i.cycles
}

type immediateMode struct {
	addressMode
}

func NewImmediate(c addressCmd, cycles uint8) instruction {
	return &immediateMode{addressMode{c, cycles}}
}

func (i *immediateMode) Execute(s *cpu.State) {
	addr := s.ProgramCounter + 1
	s.ProgramCounter += 2
	i.cmd(s, addr)
}

type zeroPageMode struct {
	addressMode
}

func NewZeroPage(c addressCmd, cycles uint8) instruction {
	return &zeroPageMode{addressMode{c, cycles}}
}

func (z *zeroPageMode) Execute(s *cpu.State) {
	addr := s.Read(s.ProgramCounter + 1)
	s.ProgramCounter += 2
	z.cmd(s, uint16(addr))
}

type zeroPageXMode struct {
	addressMode
}

func NewZeroPageX(c addressCmd, cycles uint8) instruction {
	return &zeroPageXMode{addressMode{c, cycles}}
}

func (z *zeroPageXMode) Execute(s *cpu.State) {
	addr := s.Read(s.ProgramCounter+1) + s.RegisterX
	s.ProgramCounter += 2
	z.cmd(s, uint16(addr))
}

type zeroPageYMode struct {
	addressMode
}

func NewZeroPageY(c addressCmd, cycles uint8) instruction {
	return &zeroPageYMode{addressMode{c, cycles}}
}

func (z *zeroPageYMode) Execute(s *cpu.State) {
	addr := s.Read(s.ProgramCounter+1) + s.RegisterY
	s.ProgramCounter += 2
	z.cmd(s, uint16(addr))
}

type absoluteMode struct {
	addressMode
}

func NewAbsolute(c addressCmd, cycles uint8) instruction {
	return &absoluteMode{addressMode{c, cycles}}
}

func (a *absoluteMode) Execute(s *cpu.State) {
	addr := readTwoBytes(s, s.ProgramCounter+1)
	s.ProgramCounter += 3
	a.cmd(s, addr)
}

type absoluteXMode struct {
	pageCrossMode
}

func NewAbsoluteX(c addressCmd,
	cycles, pageCrossCycles uint8) instruction {
	return &absoluteXMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *absoluteXMode) Execute(s *cpu.State) {
	baseAddr := readTwoBytes(s, s.ProgramCounter+1)
	finalAddr := baseAddr + uint16(s.RegisterX)
	a.isPageCrossed = isPageCrossed(baseAddr, finalAddr)
	s.ProgramCounter += 3
	a.cmd(s, finalAddr)
}

type absoluteYMode struct {
	pageCrossMode
}

func NewAbsoluteY(c addressCmd,
	cycles, pageCrossCycles uint8) instruction {
	return &absoluteYMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *absoluteYMode) Execute(s *cpu.State) {
	baseAddr := readTwoBytes(s, s.ProgramCounter+1)
	finalAddr := baseAddr + uint16(s.RegisterY)
	a.isPageCrossed = isPageCrossed(baseAddr, finalAddr)
	s.ProgramCounter += 3
	a.cmd(s, finalAddr)
}

type indirectMode struct {
	addressMode
}

func NewIndirect(c addressCmd, cycles uint8) instruction {
	return &indirectMode{addressMode{c, cycles}}
}

// CPU bug: https://everything2.com/title/6502+indirect+JMP+bug

func (a *indirectMode) Execute(s *cpu.State) {
	pointer := readTwoBytes(s, s.ProgramCounter+1)
	addr := readTwoBytesWithCPUBug(s, pointer)
	s.ProgramCounter += 3
	a.cmd(s, addr)
}

type indirectXMode struct {
	addressMode
}

func NewIndirectX(c addressCmd, cycles uint8) instruction {
	return &indirectXMode{addressMode{c, cycles}}
}

func (a *indirectXMode) Execute(s *cpu.State) {
	pointer := s.Read(s.ProgramCounter+1) + s.RegisterX
	addr := readTwoBytesWithPageOverflow(s, uint16(pointer))
	s.ProgramCounter += 2
	a.cmd(s, addr)
}

type indirectYMode struct {
	pageCrossMode
}

func NewIndirectY(c addressCmd,
	cycles, pageCrossCycles uint8) instruction {
	return &indirectYMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *indirectYMode) Execute(s *cpu.State) {
	pointer := uint16(s.Read(s.ProgramCounter + 1))
	baseAddr := readTwoBytesWithPageOverflow(s, pointer)
	finalAddr := baseAddr + uint16(s.RegisterY)
	a.isPageCrossed = isPageCrossed(baseAddr, finalAddr)
	s.ProgramCounter += 2
	a.cmd(s, finalAddr)
}

type pageCrossMode struct {
	addressMode
	bonusCycles   uint8
	isPageCrossed bool
}

func newPageCrossMode(c addressCmd,
	cycles, bonus uint8) pageCrossMode {
	return pageCrossMode{
		addressMode: addressMode{c, cycles},
		bonusCycles: bonus}
}

func (p *pageCrossMode) GetCycles() uint8 {
	if p.isPageCrossed {
		return p.cycles + p.bonusCycles
	}
	return p.cycles
}

type addressMode struct {
	cmd    addressCmd
	cycles uint8
}

func (a *addressMode) GetCycles() uint8 {
	return a.cycles
}

func readTwoBytes(s *cpu.State, addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(addr + 1)
	return mergeBytes(hi, lo)
}

func readTwoBytesWithCPUBug(
	s *cpu.State, addr uint16) uint16 {
	return readTwoBytesWithPageOverflow(s, addr)
}

func readTwoBytesWithPageOverflow(
	s *cpu.State, addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(addr&0xff00 + (addr+1)&0x00ff)
	return mergeBytes(hi, lo)
}

func mergeBytes(hi, lo byte) uint16 {
	return uint16(hi)<<8 + uint16(lo)
}

type relativeMode struct {
	cmd         relativeCmd
	bonusCycles uint8
}

func NewRelative(c relativeCmd) instruction {
	return &relativeMode{c, 0}
}

func (r *relativeMode) Execute(s *cpu.State) {
	r.bonusCycles = 0
	shift := convertToUint16(s.Read(s.ProgramCounter + 1))
	s.ProgramCounter += 2
	r.runCmd(s, shift)
}

func convertToUint16(shift byte) uint16 {
	s := uint16(shift)
	if isNegative(shift) {
		s |= 0xff00
	}
	return s
}

func isNegative(b byte) bool {
	return b&0x80 == 0x80
}

func (r *relativeMode) runCmd(s *cpu.State, shift uint16) {
	if r.cmd(s.Status) {
		finalAddr := s.ProgramCounter + shift
		r.updateCycles(s.ProgramCounter, finalAddr)
		s.ProgramCounter = finalAddr
	}
}

func (r *relativeMode) updateCycles(addr1, addr2 uint16) {
	r.bonusCycles++
	if isPageCrossed(addr1, addr2) {
		r.bonusCycles++
	}
}

func (r *relativeMode) GetCycles() uint8 {
	return 2 + r.bonusCycles
}

func isPageCrossed(addr1, addr2 uint16) bool {
	return addr1&0xff00 != addr2&0xff00
}
