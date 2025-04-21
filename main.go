package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/MONDERASDOR/SaverMc/protocol"
	"github.com/MONDERASDOR/SaverMc/world"
	"github.com/MONDERASDOR/SaverMc/player"
)

const (
	ServerPort = 25565
)

type ServerStatus struct {
	Version     VersionInfo `json:"version"`
	Players     PlayersInfo `json:"players"`
	Description Chat        `json:"description"`
}

type VersionInfo struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type PlayersInfo struct {
	Max    int `json:"max"`
	Online int `json:"online"`
}

type Chat struct {
	Text string `json:"text"`
}

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:25565") // Bind to all interfaces for public IP
	if err != nil {
		log.Fatalf("Failed to bind: %v", err)
	}
	log.Printf("SaverMC listening on :%d", ServerPort)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	var state = 0 // 0 = handshake, 1 = status, 2 = login
	var username string
	for {
		_, err := protocol.ReadVarInt(conn)
		if err != nil {
			return
		}
		packetID, err := protocol.ReadVarInt(conn)
		if err != nil {
			return
		}
		if state == 0 && packetID == 0x00 {
			// Handshake
			_, _ = protocol.ReadVarInt(conn) // protocol version
			_, _ = protocol.ReadString(conn) // server address
			_, _ = protocol.ReadUnsignedShort(conn) // server port
			nextState, _ := protocol.ReadVarInt(conn)
			if nextState == 1 {
				state = 1 // status
			} else if nextState == 2 {
				state = 2 // login
			}
			continue
		}
		if state == 1 && packetID == 0x00 {
			// Status request
			status := ServerStatus{
				Version: VersionInfo{
					Name:     "SaverMC 1.20.4",
					Protocol: 765,
				},
				Players: PlayersInfo{
					Max:    20,
					Online: 0,
				},
				Description: Chat{Text: "\u00a7bSaverMC: Go Minecraft Server"},
			}
			b, _ := json.Marshal(status)
			protocol.WritePacket(conn, 0x00, protocol.WriteString(string(b)))
			continue
		}
		if state == 1 && packetID == 0x01 {
			// Ping
			payload := make([]byte, 8)
			io.ReadFull(conn, payload)
			protocol.WritePacket(conn, 0x01, payload)
			return
		}
		if state == 2 && packetID == 0x00 {
			// Login Start
			username, _ = protocol.ReadString(conn)
			entityID := int32(time.Now().UnixNano() & 0x7fffffff)
			pl := player.Player{
				UUID:     fmt.Sprintf("offline-%s", username),
				Name:     username,
				EntityID: entityID,
				X:        0,
				Y:        65,
				Z:        0,
			}
			// Send Login Success
			uuid := pl.UUID
			name := pl.Name
			data := protocol.WriteString(uuid)
			data = append(data, protocol.WriteString(name)...)
			protocol.WritePacket(conn, 0x02, data)

			// Send Join Game (simplified)
			joinData := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
			protocol.WritePacket(conn, 0x26, joinData)

			// Send initial chunk (flat terrain)
			chunk := world.GenerateChunk(0, 0)
			chunkData := world.ChunkPacketData(chunk, 0, 0)
			protocol.WritePacket(conn, 0x22, chunkData)

			// Send spawn position (0,65,0)
			spawn := []byte{0x00, 0x00, 0x41, 0x00, 0x00, 0x00, 0x00}
			protocol.WritePacket(conn, 0x4D, spawn)

			// Send player position and look
			pos := make([]byte, 25)
			protocol.WritePacket(conn, 0x39, pos)

			state = 3 // in game
			continue
		}
	}
}

// Utility functions for Minecraft protocol
func readVarInt(r io.Reader) (int, error) {
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

func readUnsignedShort(r io.Reader) (uint16, error) {
	var b [2]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

func readString(r io.Reader) (string, error) {
	strlen, err := readVarInt(r)
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

func writeVarInt(val int) []byte {
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

func writeString(s string) []byte {
	b := writeVarInt(len(s))
	b = append(b, []byte(s)...)
	return b
}

func writePacket(w io.Writer, id byte, data []byte) {
	packet := append([]byte{id}, data...)
	length := writeVarInt(len(packet))
	w.Write(length)
	w.Write(packet)
}
