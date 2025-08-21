package control

import (
	"errors"
	"testdata/structural"
)

// EdgeTypeReturns = "returns"
// EdgeTypeUses = "uses"
// EdgeTypeConstructs = "constructs"

type Service struct {
	name string
}

// Constructor pattern - should create constructs edge
func NewService(name string) *Service {
	return &Service{
		name: name,
	}
}

// Multiple constructors
func NewServiceWithDefaults() *Service {
	return &Service{
		name: "default",
	}
}

// Returns edge - functions returning specific types
func GetUser(id int) *structural.User {
	return &structural.User{ID: id}
}

func GetError() error {
	return errors.New("test error")
}

// Multiple return types
func GetUserWithError(id int) (*structural.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return &structural.User{ID: id}, nil
}

// Uses edge - function using types
func ProcessUser(u *structural.User) error {
	// Uses User type
	if u == nil {
		return errors.New("nil user")
	}
	
	// Uses Service type
	svc := &Service{}
	svc.Process()
	
	return nil
}

func (s *Service) Process() {
	// Service process method
}

// Method using other types
func (s *Service) HandleUser(u *structural.User) {
	// Service uses User
	_ = u.GetUsername()
}

type Result struct {
	Data  interface{}
	Error error
}

// Complex returns
func Execute() Result {
	return Result{
		Data:  "success",
		Error: nil,
	}
}