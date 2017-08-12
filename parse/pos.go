package parse

type Position struct {
	Line   int
	Column int
}

type Pos interface {
	Position() Position
	SetPosition(Position)
}

// PosImpl provies commonly implementations for Pos.
type PosImpl struct {
	pos Position
}

// Position return the position of the expression or statement.
func (x *PosImpl) Position() Position {
	return x.pos
}

// SetPosition is a function to specify position of the expression or statement.
func (x *PosImpl) SetPosition(pos Position) {
	x.pos = pos
}
