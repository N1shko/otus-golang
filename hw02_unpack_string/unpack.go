package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	r := []rune(s)
	var b strings.Builder
	for i, v := range r {
		fmt.Println(i, " ", string(v))
		switch {
		case i == 0 && unicode.IsDigit(v):
			return "", ErrInvalidString
		case unicode.IsDigit(v) && unicode.IsDigit(r[i-1]):
			return "", ErrInvalidString
		case i == len(r)-1 || unicode.IsDigit(v):
			if !unicode.IsDigit(v) {
				b.Write([]byte(string(v)))
			} else {
				num, err := strconv.Atoi(string(v))
				if err != nil {
					return "", ErrInvalidString
				}
				if num != 0 {
					b.Write([]byte(strings.Repeat(string(r[i-1]), num)))
				}
			}
		case !unicode.IsDigit(v) && !unicode.IsDigit(r[i+1]):
			b.Write([]byte(string(v)))
		}
	}
	return b.String(), nil
}
