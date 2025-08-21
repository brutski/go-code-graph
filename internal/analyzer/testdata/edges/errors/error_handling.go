package errors

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// EdgeTypeHandlesError = "handles_error"
// EdgeTypeWrapsError = "wraps_error"

// Basic error handling
func HandleError() error {
	err := doSomething()
	if err != nil {
		// Handles error
		return err
	}
	return nil
}

// Error wrapping with fmt.Errorf
func WrapError() error {
	err := doSomething()
	if err != nil {
		// Wraps error
		return fmt.Errorf("failed to do something: %w", err)
	}
	return nil
}

// Multiple error handling
func MultipleErrorHandling() error {
	// First error check
	err := doSomething()
	if err != nil {
		return fmt.Errorf("step 1 failed: %w", err)
	}
	
	// Second error check
	err = doSomethingElse()
	if err != nil {
		return fmt.Errorf("step 2 failed: %w", err)
	}
	
	return nil
}

// Error handling with cleanup
func HandleWithCleanup() error {
	f, err := os.Open("file.txt")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	
	data := make([]byte, 100)
	_, err = f.Read(data)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	return nil
}

// Custom error types
type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func ReturnCustomError() error {
	return &CustomError{
		Code:    404,
		Message: "not found",
	}
}

type Service struct {
	name string
}

// Error checking in methods
func (s *Service) ProcessWithError() error {
	if s.name == "" {
		return errors.New("service name is empty")
	}
	
	err := s.validate()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	return nil
}

func (s *Service) validate() error {
	return nil
}

func doSomething() error {
	return errors.New("something went wrong")
}

func doSomethingElse() error {
	return nil
}