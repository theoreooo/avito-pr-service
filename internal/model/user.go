package model

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

func NewUser(id, username, team string, active bool) *User {
	return &User{
		ID:       id,
		Username: username,
		TeamName: team,
		IsActive: active,
	}
}
