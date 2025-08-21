package basic

import "fmt"

// EdgeTypeCalls = "calls"
// Tests function/method calls

func Caller() {
	// Direct function call
	callee()
	
	// Method call
	s := &Service{}
	s.Process()
	
	// Interface method call
	var w Writer = &FileWriter{}
	w.Write([]byte("test"))
	
	// Package function call
	fmt.Println("test")
}

func callee() {
	// Target of direct call
}

type Service struct{}

func (s *Service) Process() {
	// Target of method call
}

type Writer interface {
	Write([]byte) error
}

type FileWriter struct{}

func (fw *FileWriter) Write(data []byte) error {
	// Target of interface method call
	return nil
}