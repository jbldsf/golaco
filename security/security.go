package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"golaco/model"
	"io"
)

const keyString = "52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c649"

func Encrypt(stringToEncrypt string) model.Callback {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	plaintext := []byte(stringToEncrypt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	AEAD, err := cipher.NewGCM(block)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	nonce := make([]byte, AEAD.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   200,
		Result: AEAD.Seal(nonce, nonce, plaintext, nil),
		Err:    nil,
	}
}

func Decrypt(encryptedString string) model.Callback {
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	AEAD, err := cipher.NewGCM(block)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	nonceSize := AEAD.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := AEAD.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return model.Callback{
			Code:   500,
			Result: nil,
			Err:    err,
		}
	}
	return model.Callback{
		Code:   200,
		Result: plaintext,
		Err:    nil,
	}
}
