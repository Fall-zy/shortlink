package utils

import "strings"

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeBase62(id uint64) string {
	if id == 0 {
		return string(base62Chars[0])
	}
	var sb strings.Builder
	for id > 0 {
		sb.WriteByte(base62Chars[id%62])
		id /= 62
	}
	// 反转
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func DecodeBase62(shortCode string) uint64 {
	var id uint64
	for _, c := range shortCode {
		idx := strings.IndexRune(base62Chars, c)
		if idx == -1 {
			return 0
		}
		id = id*62 + uint64(idx)
	}
	return id
}
