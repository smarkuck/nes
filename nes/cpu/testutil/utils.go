package testutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/smarkuck/nes/nes/cpu/byteutil"
)

type TestBus map[uint16]byte
type Program = []byte
type Memory = map[uint16]byte

func NewTestBus(addr uint16, p Program, m Memory) TestBus {
	bus := TestBus{}
	bus.loadProgram(addr, p)
	bus.loadMemory(m)
	return bus
}

func (t TestBus) loadProgram(addr uint16, p Program) {
	for i, v := range p {
		t[addr+uint16(i)] = v
	}
}

func (t TestBus) loadMemory(m Memory) {
	for k, v := range m {
		t[k] = v
	}
}

func (t TestBus) Read(addr uint16) byte {
	return t[addr]
}

func (t TestBus) Write(addr uint16, value byte) {
	t[addr] = value
}

func (t TestBus) String() string {
	if len(t) == 0 {
		return "empty"
	}
	return t.getString()
}

func (t TestBus) getString() string {
	e := t.getEntries()
	sort.Strings(e)
	return strings.Join(e, ", ")
}

func (t TestBus) getEntries() []string {
	format := byteutil.TwoHexBytes + ": " + byteutil.HexByte
	result := []string{}
	for k, v := range t {
		entry := fmt.Sprintf(format, k, v)
		result = append(result, entry)
	}
	return result
}
