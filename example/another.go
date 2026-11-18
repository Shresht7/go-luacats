package example

// Represents a location
type Loc struct {
	X, Y int
}

// Represents a cursor
type Cursor struct {
	Loc Loc
}

// Moves the cursor up
func (c *Cursor) Up() {}

// Moves the cursor down
func (c *Cursor) Down() {}
