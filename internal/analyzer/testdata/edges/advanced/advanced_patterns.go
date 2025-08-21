package advanced

import (
	"fmt"
	"log"
	"os"
	"testdata/basic"
)

// EdgeTypePromotes = "promotes"
// EdgeTypeShadows = "shadows"
// EdgeTypeClosureCaptures = "closure_captures"
// EdgeTypeDefers = "defers"
// EdgeTypePanics = "panics" 
// EdgeTypeRecovers = "recovers"

// Method promotion through embedding
type LoggerBase struct{}

func (l *LoggerBase) Log(msg string) {
	fmt.Println(msg)
}

type Service struct {
	LoggerBase // Promotes Log method
	name       string
}

// Shadowing
func Shadowing() {
	x := 10 // outer x
	
	// Use outer x before shadowing
	println(x)
	
	if true {
		x := 20 // shadows outer x
		println(x)
	}
	
	// Function parameter shadowing
	fn := func(x int) { // shadows outer x
		println(x)
	}
	fn(30)
}

// Field shadowing through embedding
type BaseType struct {
	Name string
}

type DerivedType struct {
	BaseType
	Name string // shadows BaseType.Name
}

// Closure captures
func ClosureExamples() {
	message := "hello"
	counter := 0
	
	// Closure capturing variables
	fn := func() {
		println(message) // captures message
		counter++        // captures counter
	}
	fn()
	
	// Closure in goroutine
	go func() {
		fmt.Println(message) // captures message
	}()
	
	// Closure capturing loop variable (common bug)
	var funcs []func()
	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() {
			println(i) // captures loop variable
		})
	}
}

// Defer statements
func DeferExamples() error {
	// Simple defer
	defer fmt.Println("deferred")
	
	// Multiple defers (LIFO order)
	defer func() {
		log.Println("cleanup 1")
	}()
	
	defer func() {
		log.Println("cleanup 2")
	}()
	
	// Defer with error handling
	f, err := os.Create("temp.txt")
	if err != nil {
		return err
	}
	defer f.Close() // defers Close method
	
	// Defer with recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered: %v", r)
		}
	}()
	
	return nil
}

// Panic and recover
func PanicRecover() {
	// Function that might panic
	defer func() {
		if r := recover(); r != nil { // recovers from panic
			fmt.Printf("Recovered from: %v\n", r)
		}
	}()
	
	// Explicit panic
	if true {
		panic("something went wrong") // panics
	}
}

// Method with defer
func (s *Service) Cleanup() {
	defer s.Log("cleanup completed") // defers promoted method
	
	// Do work...
	
	if s.name == "" {
		panic("service name is empty") // panics
	}
}

// Complex defer with closure
func ComplexDefer() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	
	// May panic
	riskyOperation()
	
	return nil
}

func riskyOperation() {
	panic("risky operation failed")
}

// Parameter type relationships
func ParameterTypes(
	user *basic.User,     // parameter has type from another package
	handler func(string), // parameter has function type
) {
	// Function body
}

type User struct {
	Name string
}