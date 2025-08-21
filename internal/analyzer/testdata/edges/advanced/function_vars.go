package advanced

// Test function variable calls

func regularFunc() {
	println("regular function")
}

type HandlerFunc func(string) error

func TestFunctionVariables() {
	// Function variable - direct assignment
	fn := regularFunc
	fn() // This was causing the panic
	
	// Function variable with type
	var handler HandlerFunc = func(s string) error {
		println(s)
		return nil
	}
	handler("test")
	
	// Function field in struct
	type Server struct {
		onConnect func()
		handler   HandlerFunc
	}
	
	server := Server{
		onConnect: func() {
			println("connected")
		},
		handler: handler,
	}
	
	server.onConnect()
	server.handler("field call")
	
	// Function in slice
	funcs := []func(){regularFunc, fn}
	for _, f := range funcs {
		f()
	}
	
	// Function in map
	handlers := map[string]func(){
		"regular": regularFunc,
		"fn":      fn,
	}
	handlers["regular"]()
	
	// Higher order function
	callFunc := func(f func()) {
		f()
	}
	callFunc(regularFunc)
	callFunc(fn)
}

// Test method values
type Calculator struct {
	value int
}

func (c Calculator) Add(n int) int {
	return c.value + n
}

func TestMethodValues() {
	calc := Calculator{value: 10}
	
	// Method value
	addMethod := calc.Add
	result := addMethod(5) // This could also cause issues
	println(result)
	
	// Method expression
	addExpr := Calculator.Add
	result2 := addExpr(calc, 7)
	println(result2)
}