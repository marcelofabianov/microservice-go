package port

type PasswordHasherPort interface {
	Hash(password string) (string, error)
}
