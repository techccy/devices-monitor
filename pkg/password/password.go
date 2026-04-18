package password

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultCost = bcrypt.DefaultCost
)

type Hasher struct {
	cost int
}

func NewHasher(cost int) *Hasher {
	if cost <= 0 {
		cost = DefaultCost
	}
	return &Hasher{cost: cost}
}

func (h *Hasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	return string(bytes), err
}

func (h *Hasher) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
