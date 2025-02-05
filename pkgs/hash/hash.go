package hash

type Hasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) error
}
