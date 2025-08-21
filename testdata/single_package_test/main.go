package single_package_test

import "fmt"

// Global variable for shadowing test
var globalVar = 100

func main() {
    // First shadow: local variable shadows global
    globalVar := 200 // SHADOW 1
    fmt.Println(globalVar)
    
    // Call function that uses recover
    safeFunction()
    
    // Call another function with recover
    anotherSafeFunction()
}

// Function with one recover
func safeFunction() {
    defer func() {
        if r := recover(); r != nil { // RECOVER 1
            fmt.Println("Recovered:", r)
        }
    }()
    
    panic("test panic")
}

// Function with another recover  
func anotherSafeFunction() {
    defer func() {
        if err := recover(); err != nil { // RECOVER 2
            fmt.Println("Also recovered:", err)
        }
    }()
    
    // Second shadow: parameter shadows in inner scope
    processData(42)
}

func processData(value int) {
    if true {
        value := 100 // SHADOW 2 - shadows parameter
        fmt.Println(value)
    }
}