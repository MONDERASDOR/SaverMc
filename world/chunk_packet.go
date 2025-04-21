package world

import (
	"bytes"
	"github.com/MONDERASDOR/SaverMc/protocol"
)

// ChunkPacketData serializes a chunk into the format expected by the Minecraft 1.12.2 client
func ChunkPacketData(c Chunk, chunkX, chunkZ int) []byte {
	var buf bytes.Buffer

	// Write chunk X and Z (int32)
	buf.Write(protocol.WriteVarInt(chunkX))
	buf.Write(protocol.WriteVarInt(chunkZ))

	// Full chunk (1 byte, always 1 for full chunk)
	buf.WriteByte(1)

	// Primary bit mask (1 section = 0x1)
	buf.WriteByte(0x01)

	// Heightmaps and biomes are not present in 1.12.2 chunk data

	// Section data (1 section, 16x16x16 blocks)
	section := make([]byte, 16*16*16)
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				section[y*256+z*16+x] = c.Blocks[y][x][z]
			}
		}
	}
	// Write section length (VarInt)
	buf.Write(protocol.WriteVarInt(len(section)))
	buf.Write(section)

	// Block light (2048 bytes, all 0xff for max light)
	light := make([]byte, 2048)
	for i := range light {
		light[i] = 0xff
	}
	buf.Write(light)

	// Sky light (2048 bytes, all 0xff for max light)
	skylight := make([]byte, 2048)
	for i := range skylight {
		skylight[i] = 0xff
	}
	buf.Write(skylight)

	// No block entities
	buf.Write(protocol.WriteVarInt(0))

	return buf.Bytes()
}
