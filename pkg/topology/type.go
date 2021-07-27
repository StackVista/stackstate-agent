package topology

//go:generate msgp

// Type of a topology element (component or relation)
type Type struct {
	Name string `json:"name" msg:"name"`
}
