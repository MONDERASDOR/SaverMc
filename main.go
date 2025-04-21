package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unsafe"
	"bytes"

	"github.com/MONDERASDOR/SaverMc/protocol"
	"github.com/MONDERASDOR/SaverMc/world"
	"github.com/MONDERASDOR/SaverMc/player"
	"github.com/google/uuid"
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
	ln, err := net.Listen("tcp", "127.0.0.1:25565") // Localhost only
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
					Name:     "1.12.2",
					Protocol: 340,
				},
				Players: PlayersInfo{
					Max:    20,
					Online: 0,
				},
				Description: Chat{Text: "\u00a7bSaverMC: Go Minecraft Server"},
			}
			b, _ := json.Marshal(status)
			protocol.WritePacket(conn, 0, protocol.WriteString(string(b)))
			continue
		}
		if state == 1 && packetID == 0x01 {
			// Ping
			payload := make([]byte, 8)
			io.ReadFull(conn, payload)
			protocol.WritePacket(conn, 1, payload)
			return
		}
		if state == 2 && packetID == 0x00 {
			// Login Start
			username, _ = protocol.ReadString(conn)
			entityID := int32(time.Now().UnixNano() & 0x7fffffff)
			pl := player.Player{
				UUID:     OfflineUUID(username),
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
			protocol.WritePacket(conn, 2, data)

			// Send Join Game (0x23 for 1.12.2)
			joinBuf := new(bytes.Buffer)
			binary.Write(joinBuf, binary.BigEndian, entityID)           // Entity ID (int32)
			joinBuf.WriteByte(1)                                        // Gamemode (byte)
			binary.Write(joinBuf, binary.BigEndian, int32(0))           // Dimension (int32)
			joinBuf.WriteByte(0)                                        // Difficulty (byte)
			joinBuf.WriteByte(20)                                       // Max players (byte)
			joinBuf.Write(protocol.WriteString("default"))               // Level type (string, VarInt length prefix)
			joinBuf.Write([]byte{0x00})                                 // Reduced debug info (boolean, 0=false)
			protocol.WritePacket(conn, 35, joinBuf.Bytes())

			// Send initial chunk (flat terrain, 0x20 for 1.12.2)
			chunk := world.GenerateChunk(0, 0)
			chunkData := world.ChunkPacketData(chunk, 0, 0)
			protocol.WritePacket(conn, 32, chunkData)

			// Send spawn position (0x43 for 1.12.2)
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, int32(0))  // X
			binary.Write(buf, binary.BigEndian, int32(65)) // Y
			binary.Write(buf, binary.BigEndian, int32(0))  // Z
			protocol.WritePacket(conn, 67, buf.Bytes())

			// Send player position and look (0x2F for 1.12.2)
			teleportID := 1
			posPacket := writePlayerPositionAndLook(0, 65, 0, 0, 0, teleportID)
			protocol.WritePacket(conn, 47, posPacket)

			state = 3 // in game
			continue
		}
		if state == 3 {
			// In-game packet handling (movement, block interaction)
			switch packetID {
			case 0x0E: // Player Position (1.12.2)
				// Read position (double x, y, z, onGround bool)
				var pos [3]float64
				for i := 0; i < 3; i++ {
					b := make([]byte, 8)
					io.ReadFull(conn, b)
					pos[i] = float64FromBytes(b)
				}
				io.ReadFull(conn, make([]byte, 1)) // onGround
				// (Update player position here if desired)
			case 0x0F: // Player Position and Look (1.12.2)
				var pos [3]float64
				for i := 0; i < 3; i++ {
					b := make([]byte, 8)
					io.ReadFull(conn, b)
					pos[i] = float64FromBytes(b)
				}
				io.ReadFull(conn, make([]byte, 4)) // yaw, pitch
				io.ReadFull(conn, make([]byte, 1)) // onGround
				// (Update player position here if desired)
			case 0x13: // Player Digging (block breaking, 1.12.2)
				// Read block position (int x, y, z)
				b := make([]byte, 12)
				io.ReadFull(conn, b)
				x := int(int32FromBytes(b[0:4]))
				y := int(int32FromBytes(b[4:8]))
				z := int(int32FromBytes(b[8:12]))
				// Update chunk in memory (set to air=0)
				if y >= 0 && y < world.ChunkHeight && x >= 0 && x < world.ChunkSize && z >= 0 && z < world.ChunkSize {
					chunk := world.GenerateChunk(0, 0)
					chunk.Blocks[y][x][z] = 0
					change := make([]byte, 8)
					copy(change[0:4], int32ToBytes(int32(x)))
					copy(change[4:8], int32ToBytes(int32(z)))
					protocol.WritePacket(conn, 11, change) // Block Change (1.12.2)
				}
			case 0x14: // Block Placement (1.12.2)
				// (Stub: Accept and ignore for now)
				io.ReadFull(conn, make([]byte, 12)) // Block position
				// (Expand: Update chunk and send block change)
			}
			continue
		}
	}
}

// Utility: Convert 8 bytes to float64
func float64FromBytes(b []byte) float64 {
	bits := uint64(0)
	for i := 0; i < 8; i++ {
		bits |= uint64(b[i]) << (8 * (7 - i))
	}
	return float64FromBits(bits)
}

func float64FromBits(bits uint64) float64 {
	return *(*float64)(unsafe.Pointer(&bits))
}

// Utility: Convert 4 bytes to int32
func int32FromBytes(b []byte) int32 {
	return int32(b[0])<<24 | int32(b[1])<<16 | int32(b[2])<<8 | int32(b[3])
}

func int32ToBytes(i int32) []byte {
	return []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
}

// OfflineUUID generates a valid offline-mode UUID for a username
func OfflineUUID(username string) string {
	data := []byte("OfflinePlayer:" + username)
	hash := md5.Sum(data)
	hash[6] = (hash[6] & 0x0f) | 0x30 // version 3
	hash[8] = (hash[8] & 0x3f) | 0x80 // variant is 10
	return uuid.UUID(hash).String()
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

func writePacket(w io.Writer, id int, data []byte) {
	packet := append(writeVarInt(id), data...)
	length := writeVarInt(len(packet))
	w.Write(length)
	w.Write(packet)
}

// writePlayerPositionAndLook creates the correct 1.12.2 Player Position and Look packet
func writePlayerPositionAndLook(x, y, z float64, yaw, pitch float32, teleportID int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, x)
	binary.Write(buf, binary.BigEndian, y)
	binary.Write(buf, binary.BigEndian, z)
	binary.Write(buf, binary.BigEndian, yaw)
	binary.Write(buf, binary.BigEndian, pitch)
	buf.WriteByte(0x00) // Flags: 0 = absolute position
	buf.Write(protocol.WriteVarInt(teleportID))
	return buf.Bytes()
}
