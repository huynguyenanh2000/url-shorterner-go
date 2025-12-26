package base62

import "strings"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Encode(num uint64) string {
	if num == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder
	for num > 0 {
		sb.WriteByte(alphabet[num%62])
		num /= 62
	}

	res := sb.String()
	runes := []rune(res)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
