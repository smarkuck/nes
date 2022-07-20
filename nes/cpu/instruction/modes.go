package instruction

import "github.com/smarkuck/nes/nes/cpu/byteutil"

const (
	impliedInstrSize      = 1
	immediateInstrSize    = 2
	oneByteAddrInstrSize  = 2
	twoBytesAddrInstrSize = 3
	relativeInstrSize     = 2
	relativeInstrCycles   = 2
)

type impliedCmd = func(*State)
type addressCmd = func(_ *State, addr uint16)
type relativeCmd = func(status byte) bool

type impliedMode struct {
	cmd    impliedCmd
	cycles uint8
}

func NewImplied(c impliedCmd, cycles uint8) Instruction {
	return &impliedMode{c, cycles}
}

func NewAccumulative(c impliedCmd, cycles uint8) Instruction {
	return &impliedMode{c, cycles}
}

func (i *impliedMode) Execute(s *State) {
	s.ProgramCounter += impliedInstrSize
	i.cmd(s)
}

func (i *impliedMode) GetCycles() uint8 {
	return i.cycles
}

type immediateMode struct {
	addressMode
}

func NewImmediate(c addressCmd, cycles uint8) Instruction {
	return &immediateMode{addressMode{c, cycles}}
}

func (i *immediateMode) Execute(s *State) {
	addr := s.GetParamAddress()
	s.ProgramCounter += immediateInstrSize
	i.cmd(s, addr)
}

type zeroPageMode struct {
	addressMode
}

func NewZeroPage(c addressCmd, cycles uint8) Instruction {
	return &zeroPageMode{addressMode{c, cycles}}
}

func (z *zeroPageMode) Execute(s *State) {
	addr := s.ReadOneByteParam()
	s.ProgramCounter += oneByteAddrInstrSize
	z.cmd(s, uint16(addr))
}

type zeroPageXMode struct {
	addressMode
}

func NewZeroPageX(c addressCmd, cycles uint8) Instruction {
	return &zeroPageXMode{addressMode{c, cycles}}
}

func (z *zeroPageXMode) Execute(s *State) {
	addr := s.ReadOneByteParam() + s.RegisterX
	s.ProgramCounter += oneByteAddrInstrSize
	z.cmd(s, uint16(addr))
}

type zeroPageYMode struct {
	addressMode
}

func NewZeroPageY(c addressCmd, cycles uint8) Instruction {
	return &zeroPageYMode{addressMode{c, cycles}}
}

func (z *zeroPageYMode) Execute(s *State) {
	addr := s.ReadOneByteParam() + s.RegisterY
	s.ProgramCounter += oneByteAddrInstrSize
	z.cmd(s, uint16(addr))
}

type absoluteMode struct {
	addressMode
}

func NewAbsolute(c addressCmd, cycles uint8) Instruction {
	return &absoluteMode{addressMode{c, cycles}}
}

func (a *absoluteMode) Execute(s *State) {
	addr := s.ReadTwoBytesParam()
	s.ProgramCounter += twoBytesAddrInstrSize
	a.cmd(s, addr)
}

type absoluteXMode struct {
	pageCrossMode
}

func NewAbsoluteX(c addressCmd,
	cycles, pageCrossCycles uint8) Instruction {
	return &absoluteXMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *absoluteXMode) Execute(s *State) {
	base := s.ReadTwoBytesParam()
	final := base + uint16(s.RegisterX)
	a.checkPageCross(base, final)
	s.ProgramCounter += twoBytesAddrInstrSize
	a.cmd(s, final)
}

type absoluteYMode struct {
	pageCrossMode
}

func NewAbsoluteY(c addressCmd,
	cycles, pageCrossCycles uint8) Instruction {
	return &absoluteYMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *absoluteYMode) Execute(s *State) {
	base := s.ReadTwoBytesParam()
	final := base + uint16(s.RegisterY)
	a.checkPageCross(base, final)
	s.ProgramCounter += twoBytesAddrInstrSize
	a.cmd(s, final)
}

type indirectMode struct {
	addressMode
}

func NewIndirect(c addressCmd, cycles uint8) Instruction {
	return &indirectMode{addressMode{c, cycles}}
}

func (a *indirectMode) Execute(s *State) {
	pointer := s.ReadTwoBytesParam()
	// CPU bug: https://everything2.com/title/6502+indirect+JMP+bug
	addr := s.ReadTwoBytesPageOverflow(pointer)
	s.ProgramCounter += twoBytesAddrInstrSize
	a.cmd(s, addr)
}

type indirectXMode struct {
	addressMode
}

func NewIndirectX(c addressCmd, cycles uint8) Instruction {
	return &indirectXMode{addressMode{c, cycles}}
}

func (a *indirectXMode) Execute(s *State) {
	pointer := s.ReadOneByteParam() + s.RegisterX
	addr := s.ReadTwoBytesPageOverflow(uint16(pointer))
	s.ProgramCounter += oneByteAddrInstrSize
	a.cmd(s, addr)
}

type indirectYMode struct {
	pageCrossMode
}

func NewIndirectY(c addressCmd,
	cycles, pageCrossCycles uint8) Instruction {
	return &indirectYMode{
		newPageCrossMode(c, cycles, pageCrossCycles)}
}

func (a *indirectYMode) Execute(s *State) {
	base, final := a.getAddresses(s)
	a.checkPageCross(base, final)
	s.ProgramCounter += oneByteAddrInstrSize
	a.cmd(s, final)
}

func (a *indirectYMode) getAddresses(
	s *State) (uint16, uint16) {
	pointer := uint16(s.ReadOneByteParam())
	base := s.ReadTwoBytesPageOverflow(pointer)
	final := base + uint16(s.RegisterY)
	return base, final
}

type pageCrossMode struct {
	addressMode
	bonusCycles uint8
	isPageCross bool
}

func newPageCrossMode(c addressCmd,
	cycles, bonus uint8) pageCrossMode {
	return pageCrossMode{
		addressMode: addressMode{c, cycles},
		bonusCycles: bonus}
}

func (p *pageCrossMode) checkPageCross(
	base, final uint16) {
	p.isPageCross = !byteutil.IsSameHighByte(base, final)
}

func (p *pageCrossMode) GetCycles() uint8 {
	if p.isPageCross {
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

type relativeMode struct {
	cmd         relativeCmd
	bonusCycles uint8
}

func NewRelative(c relativeCmd) Instruction {
	return &relativeMode{c, 0}
}

func (r *relativeMode) Execute(s *State) {
	r.bonusCycles = 0
	shift := s.ReadOneByteParam()
	shift16 := byteutil.ToArithmeticUint16(shift)
	s.ProgramCounter += relativeInstrSize
	r.runCmd(s, shift16)
}

func (r *relativeMode) runCmd(s *State, shift uint16) {
	if r.cmd(s.Status) {
		finalAddr := s.ProgramCounter + shift
		r.updateCycles(s.ProgramCounter, finalAddr)
		s.ProgramCounter = finalAddr
	}
}

func (r *relativeMode) updateCycles(base, final uint16) {
	r.bonusCycles++
	if !byteutil.IsSameHighByte(base, final) {
		r.bonusCycles++
	}
}

func (r *relativeMode) GetCycles() uint8 {
	return relativeInstrCycles + r.bonusCycles
}
