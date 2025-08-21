package basic

// EdgeTypeEmbeds = "embeds"
// Tests struct embedding

type Base struct {
	ID   string
	Name string
}

func (b *Base) GetID() string {
	return b.ID
}

// Direct embedding
type Extended struct {
	Base // Embedded struct
	Extra string
}

// Pointer embedding
type ExtendedPtr struct {
	*Base // Embedded pointer
	Extra string
}

// Interface embedding
type Logger interface {
	Log(string)
}

type ErrorLogger interface {
	Logger // Embedded interface
	Error(string)
}

// Multiple embedding
type MultiEmbed struct {
	Base
	*Service
	Extra string
}