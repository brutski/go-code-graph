package generics

// EdgeTypeInstantiates = "instantiates"
// EdgeTypeConstrains = "constrains"
// Tests generic type relationships

// Generic type with constraint
type Container[T any] struct {
	items []T
}

// Method on generic type
func (c *Container[T]) Add(item T) {
	c.items = append(c.items, item)
}

// Generic function with constraints
func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Type constraint
type Ordered interface {
	~int | ~float64 | ~string
}

// Generic instantiation
func UseGenerics() {
	// Instantiates Container[int]
	intContainer := &Container[int]{}
	intContainer.Add(42)
	
	// Instantiates Container[string]
	stringContainer := Container[string]{
		items: []string{"hello", "world"},
	}
	_ = stringContainer
	
	// Uses generic function
	maxInt := Max(10, 20)
	maxString := Max("apple", "banana")
	
	_ = maxInt
	_ = maxString
}

// Generic interface
type Processor[T any] interface {
	Process(T) error
}

// Implementation of generic interface
type IntProcessor struct{}

func (p *IntProcessor) Process(i int) error {
	return nil
}

// Generic with multiple type parameters
type Pair[T, U any] struct {
	First  T
	Second U
}

// Generic function with multiple constraints
func Transform[T, U any](input T, fn func(T) U) U {
	return fn(input)
}

// Type with embedded constraint
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](values []T) T {
	var sum T
	for _, v := range values {
		sum += v
	}
	return sum
}