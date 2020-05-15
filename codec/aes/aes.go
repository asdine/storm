package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/asdine/storm/v3/codec"
)

const name = "aes-"

// AES is an Codec that encrypts the data and uses a sub marshaller to actually serialize the data
type AES struct {
	subMarshaller codec.MarshalUnmarshaler
	aesGCM        cipher.AEAD
}

// NewAES creates a new AES encryption marshaller. It takes a sub marshaller to actually serialize the data and a 16/32 bytes private key to
// encrypt all data using AES in GCM block mode.
func NewAES(subMarshaller codec.MarshalUnmarshaler, key []byte) (*AES, error) {
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, fmt.Errorf("error creating GCM block mode: %w", err)
	}

	return &AES{
		subMarshaller: subMarshaller,
		aesGCM:        aesGCM,
	}, nil
}

func (c *AES) Name() string {
	// Return a dynamic name, because the marshalling will also fail if the inner marshalling changes.
	return name + c.subMarshaller.Name()
}

func (c *AES) Marshal(v interface{}) ([]byte, error) {
	data, err := c.subMarshaller.Marshal(v)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, c.aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("error reading random nonce: %w", err)
	}

	return c.aesGCM.Seal(nonce, nonce, data, nil), nil
}

func (c *AES) Unmarshal(data []byte, v interface{}) error {
	nonceSize := c.aesGCM.NonceSize()
	decrypted, err := c.aesGCM.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return fmt.Errorf("error decrypting data: %w", err)
	}

	return c.subMarshaller.Unmarshal(decrypted, v)
}
