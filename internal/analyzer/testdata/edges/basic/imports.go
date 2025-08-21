package basic

// EdgeTypeImports = "imports"
// Tests package imports

import (
	"fmt"      // Standard library import
	"net/http" // Another std import
	
	// Aliased import
	ioutil "io"
	
	// Blank import for side effects
	_ "image/png"
)

func UseImports() {
	fmt.Println("using fmt")
	http.Get("https://example.com")
	ioutil.ReadAll(nil)
}