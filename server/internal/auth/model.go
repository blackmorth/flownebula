package auth

type User struct {
    ID           int64    `json:"id" db:"id"`
    Email        string   `json:"email" db:"email"`
    Password     string   `json:"-" db:"password"` // hashé
    Roles        []string `json:"roles" db:"roles"` // JSON
    AgentToken   string   `json:"agent_token" db:"agent_token"`
    AgentEnabled bool     `json:"agent_enabled" db:"agent_enabled"`
}
