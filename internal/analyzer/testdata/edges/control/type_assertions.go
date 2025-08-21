package control

import (
	"errors"
	"testdata/basic"
	"testdata/structural"
)

// EdgeTypeTypeAsserts = "type_asserts"
// Tests type assertions and type switches

func TypeAssertions(i interface{}) {
	// Simple type assertion
	s, ok := i.(string)
	if ok {
		println(s)
	}
	
	// Direct type assertion (can panic)
	str := i.(string)
	println(str)
	
	// Asserting to interface
	w, ok := i.(basic.Writer)
	if ok {
		w.Write([]byte("test"))
	}
}

func TypeSwitch(i interface{}) string {
	// Type switch
	switch v := i.(type) {
	case string:
		return "string: " + v
	case int:
		return "int"
	case *structural.User:
		return "user: " + v.Username
	case basic.Writer:
		return "writer"
	default:
		return "unknown"
	}
}

// Generic type assertion
func AssertType[T any](v interface{}) (T, bool) {
	t, ok := v.(T)
	return t, ok
}

// Method with type assertion
func (s *Service) ProcessInterface(data interface{}) error {
	// Assert to specific struct
	if user, ok := data.(*structural.User); ok {
		s.HandleUser(user)
		return nil
	}
	
	// Assert to interface
	if w, ok := data.(basic.Writer); ok {
		return w.Write([]byte("data"))
	}
	
	return errors.New("unsupported type")
}