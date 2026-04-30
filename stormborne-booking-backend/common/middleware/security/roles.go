package security

type Role string

const (
	ADMIN Role = "admin"
	USER Role = "user"
	OWNER Role = "owner"
)