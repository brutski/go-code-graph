package pointer_receiver

// Test case for GitHub issue #3
// Tests that calls edges are properly constructed for pointer receiver methods
// when called by struct values (not pointers)

type FileWriter interface {
	Write([]byte) error
}

type BasicFileWriter struct {
	name string
}

// Pointer receiver method
func (f *BasicFileWriter) Write(data []byte) error {
	return nil
}

type Service struct {
	name string
}

// Pointer receiver method that takes a FileWriter interface
func (s *Service) PubackEncode(f FileWriter) error {
	// This call should create a proper edge to *BasicFileWriter.Write
	// even when f is a value, not a pointer
	err := f.Write([]byte("test"))
	if err != nil {
		return err
	}
	return nil
}

// Test function that demonstrates the issue
func TestPointerReceiverCall() {
	service := Service{name: "test"}

	// Create a value (not pointer) of BasicFileWriter
	writer := BasicFileWriter{name: "test"}

	// Call PubackEncode with a pointer to the value
	// This internally calls writer.Write() which has a pointer receiver
	service.PubackEncode(&writer)
}

// Additional test for the specific issue: value calling pointer receiver method
func TestValueCallsPointerMethod() {
	// Create a value (not pointer) of BasicFileWriter
	writer := BasicFileWriter{name: "test"}

	// Direct call on value - Go automatically takes address: (&writer).Write()
	// This is the exact scenario from the GitHub issue
	writer.Write([]byte("value"))
}

// Additional test cases for completeness
func TestDirectPointerCall() {
	// Direct call on pointer - this should work correctly
	writer := &BasicFileWriter{name: "test"}
	writer.Write([]byte("direct"))
}

func TestValueCall() {
	// Direct call on value - this should also work with the fix
	writer := BasicFileWriter{name: "test"}
	writer.Write([]byte("value"))
}