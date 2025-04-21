package world

import (
	"bytes"
	"encoding/binary"
	"github.com/MONDERASDOR/SaverMc/protocol"
)

func ChunkPacketData(c Chunk, chunkX, chunkZ int) []byte {
	var buf bytes.Buffer

	// Write chunk X and Z (int32, big endian)
	binary.Write(&buf, binary.BigEndian, int32(chunkX))
	binary.Write(&buf, binary.BigEndian, int32(chunkZ))

	// Ground-up continuous (bool, 1 byte)
	buf.WriteByte(1)

	// Build section bitmask and section data
	var bitmask uint16
	sections := make([][]byte, 0, 16)
	for sectionIdx := 0; sectionIdx < 16; sectionIdx++ {
		nonAir := false
		blockCount := 0
		sectionBlocks := make([]byte, 4096)
		idx := 0
		for y := sectionIdx * 16; y < (sectionIdx+1)*16; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					block := c.Blocks[y][x][z]
					sectionBlocks[idx] = block
					if block != 0 {
						nonAir = true
						blockCount++
					}
					idx++
				}
			}
		}
		if nonAir {
			bitmask |= 1 << uint(sectionIdx)
			section := new(bytes.Buffer)
			// Write block count (2 bytes, big endian)
			section.WriteByte(byte(blockCount >> 8))
			section.WriteByte(byte(blockCount))
			// Write block IDs
			section.Write(sectionBlocks)
			// Block light (2048 bytes, all 0xff)
			light := make([]byte, 2048)
			for i := range light { light[i] = 0xff }
			section.Write(light)
			// Sky light (2048 bytes, all 0xff)
			skylight := make([]byte, 2048)
			for i := range skylight { skylight[i] = 0xff }
			section.Write(skylight)
			sections = append(sections, section.Bytes())
		}
	}

	// Write bitmask (2 bytes, big endian)
	buf.Write([]byte{byte(bitmask >> 8), byte(bitmask)})

	// Write section data
	for _, section := range sections {
		buf.Write(protocol.WriteVarInt(len(section)))
		buf.Write(section)
	}

	// Biomes (256 bytes, all plains)
	biomes := make([]byte, 256)
	for i := range biomes { biomes[i] = 1 }
	buf.Write(biomes)

	// No block entities
	buf.Write(protocol.WriteVarInt(0))

	return buf.Bytes()
}
