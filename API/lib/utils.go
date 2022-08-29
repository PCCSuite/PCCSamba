package lib

import (
	"github.com/dchest/uniuri"
)

const passwordLength = 16

var passwordLetters = uniuri.StdChars

func GeneratePassword() string {
	return uniuri.NewLenChars(passwordLength, passwordLetters)
}
