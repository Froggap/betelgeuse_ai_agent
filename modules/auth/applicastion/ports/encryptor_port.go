package authports

type EncryptorPort interface {
	Encode(password string) (string, error)
	Compare(password string, hashedPassword string) (bool, error)
}
