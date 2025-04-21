package player

// Player represents a connected player
// For now, minimal info (expand later)
type Player struct {
	UUID     string
	Name     string
	EntityID int32
	X, Y, Z  float64
	Yaw, Pitch float32
	Inventory [36]byte // Simple inventory for now
}
