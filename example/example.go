package example

import "math"

// Position in 3D space
type Position struct {
	X float64
	Y float64
	Z float64
}

// Distance between two positions
func Distance(a, b *Position) float64 {
	dX := b.X - a.X
	dY := b.Y - a.Y
	dZ := b.Z - a.Z
	return math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2) + math.Pow(dZ, 2))
}

// Line in 3D space
type Line struct {
	start *Position
	end   *Position
}

// Length of the line
func (l *Line) Length() float64 {
	return Distance(l.start, l.end)
}
