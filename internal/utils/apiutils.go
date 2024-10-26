package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func GenerateTID() (string, error) {
	bid, err := generateID(12, 9)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("TSK-%v", bid), nil
}

func generateID(length int, mx int64) (string, error) {
	max := big.NewInt(mx)
	ilen := length
	id := ""
	errors := make([]error, 0)

	for i := 0; i <= ilen; i++ {
		n, err := rand.Int(rand.Reader, max)

		if err != nil {
			errors = append(errors, err)
		}

		id += fmt.Sprintf("%v", n)
	}

	if len(errors) > 0 {
		return "", fmt.Errorf("error occurred while generating id:: %w", errors[0])
	}

	return id, nil
}

func GenToken() ([]byte, error) {
	ilen := 32
	asci := make([]byte, ilen)
	_, err := rand.Read(asci)
	if err != nil {
		return nil, fmt.Errorf("error occurred while generating id:: %w", err)
	}
	return asci, nil
}
