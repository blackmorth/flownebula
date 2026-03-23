package auth

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id int64) (*User, error)
	FindByAgentToken(token string) (*User, error)
	FindAll() ([]*User, error)
	UpdateAgentEnabled(id int64, enabled bool) error
	UpdateAgentToken(id int64, token string) error
}
