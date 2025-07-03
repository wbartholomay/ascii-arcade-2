package vector

type Vector struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

func (vector *Vector) Add(otherVector Vector) {
	vector.X += otherVector.X
	vector.Y += otherVector.Y
}