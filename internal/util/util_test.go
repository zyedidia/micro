package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringWidth(t *testing.T) {
	bytes := []byte("\tPot să \tmănânc sticlă și ea nu mă rănește.")

	n := StringWidth(bytes, 23, 4)
	assert.Equal(t, 26, n)
}

func TestSliceVisualEnd(t *testing.T) {
	s := []byte("\thello")
	slc, n, _ := SliceVisualEnd(s, 2, 4)
	assert.Equal(t, []byte("\thello"), slc)
	assert.Equal(t, 2, n)

	slc, n, _ = SliceVisualEnd(s, 1, 4)
	assert.Equal(t, []byte("\thello"), slc)
	assert.Equal(t, 1, n)

	slc, n, _ = SliceVisualEnd(s, 4, 4)
	assert.Equal(t, []byte("hello"), slc)
	assert.Equal(t, 0, n)

	slc, n, _ = SliceVisualEnd(s, 5, 4)
	assert.Equal(t, []byte("ello"), slc)
	assert.Equal(t, 0, n)
}

func TestEncrypt(t *testing.T) {
	result, err := Encrypt("This is a test string to encrypt", "thisisasamplekey")
	assert.Nil(t, err)
	assert.NotEmpty(t, result)
}

func TestEncryptAndDecrypt(t *testing.T) {
	encryptResult, err := Encrypt("This is a test string to encrypt", "thisisasamplekey")
	assert.Nil(t, err)
	assert.NotEmpty(t, encryptResult)

	decryptResult, err := Decrypt(encryptResult, "thisisasamplekey")
	assert.Nil(t, err)
	assert.Equal(t, "This is a test string to encrypt", decryptResult)
}

func TestEncryptWithEmptyKey(t *testing.T) {
	result, err := Encrypt("This is a test string to encrypt", "")
	assert.NotNil(t, err)
	assert.Equal(t, "encryption key cannot be empty", result)
}
