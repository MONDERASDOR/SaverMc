package protocol

import (
	"io"
	"fmt"
)

// Utility functions for Minecraft protocol
func ReadVarInt(r io.Reader) (int, error) {
	var num int
	var shift uint
	for {
		var b [1]byte
		_, err := r.Read(b[:])
		if err != nil {
			return 0, err
		}
		num |= int(b[0]&0x7F) << shift
		if b[0]&0x80 == 0 {
			break
		}
		shift += 7
		if shift > 35 {
			return 0, fmt.Errorf("VarInt too big")
		}
	}
	return num, nil
}

func ReadUnsignedShort(r io.Reader) (uint16, error) {
	var b [2]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

func ReadString(r io.Reader) (string, error) {
	strlen, err := ReadVarInt(r)
	if err != nil {
		return "", err
	}
	b := make([]byte, strlen)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func WriteVarInt(val int) []byte {
	var out []byte
	for {
		b := byte(val & 0x7F)
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if val == 0 {
			break
		}
	}
	return out
}

func WriteString(s string) []byte {
	b := WriteVarInt(len(s))
	b = append(b, []byte(s)...)
	return b
}

func WritePacket(w io.Writer, id byte, data []byte) {
	packet := append([]byte{id}, data...)
	length := WriteVarInt(len(packet))
	w.Write(length)
	w.Write(packet)
}
