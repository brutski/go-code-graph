package advanced

import (
	"net/http"
	"testdata/basic"
	"testdata/structural"
)

// EdgeTypeParameterType = "parameter_type"
// EdgeTypeMethodOf = "method_of"
// EdgeTypeCreates = "creates"

// Parameter type relationships
func ParameterTypeExamples(
	user *structural.User,           // parameter has type *User
	handler func(string),            // parameter has function type  
	opts ...structural.Option,       // parameter has type Option (variadic must be last)
) {
	// Function body
}

// Interface parameter
func ProcessInterface(r basic.Reader) error {
	// Parameter r has type Reader interface
	_, err := r.Read(nil)
	return err
}

// Generic parameter type
func GenericParam[T any](value T) T {
	// Parameter has generic type T
	return value
}

// Method relationships - method_of edges should be created
type Repository struct {
	db string
}

// Methods belong to Repository
func (r *Repository) Save(user *structural.User) error {
	// Method Save belongs to Repository
	return nil
}

func (r *Repository) Find(id int) (*structural.User, error) {
	// Method Find belongs to Repository
	return nil, nil
}

func (r Repository) Count() int {
	// Value receiver method also belongs to Repository
	return 0
}

// Creates relationships (instantiation)
func CreateInstances() {
	// Creates User instance
	u1 := User{
		Name: "test",
	}
	
	// Creates User instance via new
	u2 := new(User)
	
	// Creates slice
	users := make([]*User, 0, 10)
	
	// Creates map
	userMap := make(map[int]*User)
	
	// Creates channel
	ch := make(chan *User, 5)
	
	// Use to avoid unused warnings
	_ = u1
	_ = u2
	_ = users
	_ = userMap
	_ = ch
}

// Factory method creating instances
func (r *Repository) CreateUser(name string) *User {
	// Creates and returns User instance
	return &User{
		Name: name,
	}
}

// Method creating other types
func (s *Service) CreateProcessor() *Processor {
	// Service creates Processor
	return &Processor{
		service: s,
	}
}

type Processor struct {
	service *Service
}

// Complex parameter types
type Handler func(r *http.Request) error

func RegisterHandler(
	pattern string,              // string parameter
	handler Handler,             // custom function type parameter
	middleware ...func(Handler), // variadic function parameter
) {
	// Registration logic
}