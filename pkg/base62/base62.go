package base62

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Random 生成指定长度的 base62 随机字符串。
func Random(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(fmt.Sprintf("base62: crypto/rand failed: %v", err))
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}
