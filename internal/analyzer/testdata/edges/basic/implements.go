package basic

// EdgeTypeImplements = "implements"
// Tests interface implementation

type Reader interface {
	Read([]byte) (int, error)
}

type Closer interface {
	Close() error
}

// Single interface implementation
type FileReader struct {
	path string
}

func (fr *FileReader) Read(p []byte) (int, error) {
	return 0, nil
}

// Multiple interface implementation
type ReadWriteCloser struct {
	data []byte
}

func (rc *ReadWriteCloser) Read(p []byte) (int, error) {
	return 0, nil
}

func (rc *ReadWriteCloser) Close() error {
	return nil
}

func (rc *ReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

// Embedded interface
type ReadWriter interface {
	Reader
	Writer
}

type Buffer struct{}

func (b *Buffer) Read(p []byte) (int, error) {
	return 0, nil
}

func (b *Buffer) Write(p []byte) (int, error) {
	return len(p), nil
}