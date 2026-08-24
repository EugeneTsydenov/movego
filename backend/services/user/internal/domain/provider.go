package domain

import "fmt"

type Provider string

const (
	Google   Provider = "google"
	GitHub   Provider = "github"
	Password Provider = "password"
)

func NewProvider(providerStr string) (Provider, error) {
	switch Provider(providerStr) {
	case Google, GitHub, Password:
		return Provider(providerStr), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidProvider, providerStr)
	}
}

func (p Provider) String() string {
	return string(p)
}
