package world

import (
	"bytes"
	"github.com/yourusername/mcserver/protocol"
)

// ChunkPacketData serializes a chunk into the format expected by the Minecraft client (1.20.4)
func ChunkPacketData(c Chunk, chunkX, chunkZ int) []byte {
	// For simplicity, we send a single chunk, no biomes, no block entities, no lighting
	var buf bytes.Buffer

	// Chunk X and Z
	buf.Write(protocol.WriteVarInt(chunkX))
	buf.Write(protocol.WriteVarInt(chunkZ))
	buf.WriteByte(1) // Full chunk
	buf.Write(protocol.WriteVarInt(0)) // Primary bit mask (1 section)
	buf.Write(protocol.WriteVarInt(0)) // Heightmaps (not implemented)
	// Chunk data (just one section, 16x16x16, palette)
	section := make([]byte, 16*16*16)
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				section[y*256+z*16+x] = c.Blocks[y][x][z]
			}
		}
	}
	buf.Write(protocol.WriteVarInt(len(section)))
	buf.Write(section)
	return buf.Bytes()
}
