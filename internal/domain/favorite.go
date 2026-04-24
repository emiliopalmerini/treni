package domain

// Favorite is a saved FROM→TO route, addressable by Name within a chat.
type Favorite struct {
	Name string
	From string
	To   string
}
