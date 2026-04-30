package security
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)


func GeneratePasswordResetToken() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b); 
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}