package randstr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// RandomHex generates n random bytes and returns their hex-encoded string.
// Falls back to a timestamp-based value if crypto/rand fails.
func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
