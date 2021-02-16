package utils

import (
	"io/ioutil"

	"github.com/gobuffalo/envy"
)

// ReadJWTKey - Read the content of jwt sign key
func ReadJWTKey() ([]byte, error) {
	keyPath := envy.Get("JWT_KEY_PATH", "")

	content, error := ioutil.ReadFile(keyPath)

	return content, error
}
