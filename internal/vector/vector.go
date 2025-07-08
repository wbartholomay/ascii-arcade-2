package vector

type Vector struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

func NewVector(X, Y int) Vector {
	return Vector{X: X, Y: Y}
}

func (vector *Vector) Add(otherVector Vector) {
	vector.X += otherVector.X
	vector.Y += otherVector.Y
}

func (vector *Vector) Equals(otherVector Vector) bool {
	return vector.X == otherVector.X && vector.Y == otherVector.Y
}
