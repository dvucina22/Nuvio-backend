package types

import "fmt"

type Provider string

const (
	ProviderGoogle Provider = "google"
)

func (p Provider) String() string {
	return string(p)
}

func ParseProvider(s string) (Provider, error) {
	switch s {
	case "google":
		return ProviderGoogle, nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", s)
	}
}
