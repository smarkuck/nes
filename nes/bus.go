package nes

type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}
