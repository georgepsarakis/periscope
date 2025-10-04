package pkg

import (
	"math/big"
	"strings"
)

func IntToBase62(num int) string {
	return big.NewInt(int64(num)).Text(62)
}

func Zerofill(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat("0", length-len(s)) + s
}
