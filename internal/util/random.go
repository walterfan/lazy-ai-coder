package util

import (
        "crypto/rand"
        "fmt"
        "math/big"

        "github.com/google/uuid"

)

// GenerateUUID creates a new UUID
func GenerateUUID() string {
		return uuid.New().String()
}

// GenerateRandomInt generates a random integer between min and max (inclusive)
func GenerateRandomInt(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return int(nBig.Int64()) + min, nil
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	result := make([]byte, length)
	for i := range result {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[nBig.Int64()]
	}
	return string(result), nil
}
