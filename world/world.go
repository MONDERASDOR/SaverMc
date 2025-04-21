package world

import (
	"math"
)

// ChunkSize is the width/length of a chunk in blocks
const ChunkSize = 16
const ChunkHeight = 256

// Chunk represents a single chunk of world terrain
// For simplicity, we use a 2D slice for blocks (Y, XZ)
type Chunk struct {
	Blocks [ChunkHeight][ChunkSize][ChunkSize]byte
}

// GenerateChunk generates a simple flat or noise-based chunk
func GenerateChunk(chunkX, chunkZ int) Chunk {
	var c Chunk
	// Simple flat terrain at Y=64
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			h := 64 + int(8*simpleNoise(float64(chunkX*ChunkSize+x), float64(chunkZ*ChunkSize+z)))
			for y := 0; y < h; y++ {
				if y == h-1 {
					c.Blocks[y][x][z] = 2 // grass
				} else if y > h-5 {
					c.Blocks[y][x][z] = 3 // dirt
				} else {
					c.Blocks[y][x][z] = 1 // stone
				}
			}
		}
	}
	return c
}

// simpleNoise is a fast pseudo-random terrain generator
func simpleNoise(x, z float64) float64 {
	return math.Sin(x*0.05)+math.Cos(z*0.05)
}
