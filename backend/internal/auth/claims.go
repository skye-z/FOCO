package auth

type Claims struct {
	UserID string
	Email  string
	Roles  []string
}
