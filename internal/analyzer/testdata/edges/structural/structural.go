package structural

// EdgeTypeHasMethod = "has_method"
// EdgeTypeHasField = "has_field"
// EdgeTypeHasParameter = "has_parameter"
// Tests structural relationships

type User struct {
	// Fields - should create has_field edges
	ID       int
	Username string
	Email    string
	password string // private field
	Profile  *Profile
}

type Profile struct {
	Bio      string
	Location string
}

// Methods - should create has_method edges
func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) SetPassword(password string) {
	u.password = password
}

func (u User) IsValid() bool {
	return u.ID > 0
}

// Function with parameters - should create has_parameter edges
func CreateUser(
	id int,           // parameter 1
	username string,  // parameter 2
	email string,     // parameter 3
	opts ...Option,   // variadic parameter
) (*User, error) {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

type Option func(*User)

// Method with parameters
func (u *User) UpdateProfile(bio, location string) {
	if u.Profile == nil {
		u.Profile = &Profile{}
	}
	u.Profile.Bio = bio
	u.Profile.Location = location
}