package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joijoku/PR/shared"
	"golang.org/x/crypto/bcrypt"
)

func EncryptMessage(message, kunci string) (string, string, error) {
	// key := []byte("jalanjalankesemarangbersamabapak")
	key := []byte(kunci)
	msg := []byte(message)

	var err error
	msgEncrypted := ""
	remarks := ""
	errCode := 0

	shared.Block{
		Try: func() {
			x, err := aes.NewCipher(key)
			if err != nil {
				errCode = 1 // unable not generate a new cipher
				shared.CheckErr(err)
			} else {
				cipherText := make([]byte, aes.BlockSize+len(msg))
				iv := cipherText[:aes.BlockSize]
				if _, err := io.ReadFull(rand.Reader, iv); err != nil {
					errCode = 2 // unable to do ecryption
					shared.CheckErr(err)
				} else {
					stream := cipher.NewCFBEncrypter(x, iv)
					stream.XORKeyStream(cipherText[aes.BlockSize:], msg)

					msgEncrypted = base64.StdEncoding.EncodeToString(cipherText)
				}
			}
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
			if errCode == 1 {
				remarks = "unable not generate a new cipher"
			} else if errCode == 2 {
				remarks = "unable to do ecryption"
			}
		},
	}.Do()

	return msgEncrypted, remarks, err
}

func DecryptMessage(message, kunci string) (string, string, error) {
	errCode := 0
	remarks := ""
	msgDecrypted := ""
	// key := []byte("jalanjalankesemarangbersamabapak")
	key := []byte(kunci)
	var err error

	shared.Block{
		Try: func() {
			cipherText, err := base64.StdEncoding.DecodeString(message)
			if err != nil {
				errCode = 3 // unalble to base64 decode
			} else {
				x, err := aes.NewCipher(key)
				if err != nil {
					errCode = 1 // unable to generate new chipper
				} else {
					if len(cipherText) < aes.BlockSize {
						errCode = 4 // invalid ciphertext blocksize
					} else {
						iv := cipherText[:aes.BlockSize]
						cipherText = cipherText[aes.BlockSize:]

						stream := cipher.NewCFBDecrypter(x, iv)
						stream.XORKeyStream(cipherText, cipherText)

						msgDecrypted = string(cipherText)
					}
				}
			}
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
			if errCode == 1 {
				remarks = "unable not generate a new cipher"
			} else if errCode == 2 {
				remarks = "unable to do ecryption"
			} else if errCode == 3 {
				remarks = "unable to base64 decode"
			} else if errCode == 4 {
				remarks = "invalid ciphertext blocksize"
			}
		},
	}.Do()

	return msgDecrypted, remarks, err
}

func TripleEcbDesDecrypt(crypted, key []byte) ([]byte, error) {
	tkey := make([]byte, 24)
	copy(tkey, key)
	k1 := tkey[:8]
	k2 := tkey[8:16]
	k3 := tkey[16:]
	buf1, err := decrypt(crypted, k3)
	if err != nil {
		return nil, err
	}
	buf2, err := encrypt(buf1, k2)
	if err != nil {
		return nil, err
	}
	out, err := decrypt(buf2, k1)
	if err != nil {
		return nil, err
	}
	out = PKCS5Unpadding(out)
	return out, nil
}

// ECB PKCS5Padding
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// ECB PKCS5Unpadding
func PKCS5Unpadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// Des encryption
func encrypt(origData, key []byte) ([]byte, error) {
	if len(origData) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(origData)%bs != 0 {
		return nil, errors.New("wrong padding")
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

// Des Decrypt
func decrypt(crypted, key []byte) ([]byte, error) {
	if len(crypted) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(crypted))
	dst := out
	bs := block.BlockSize()
	if len(crypted)%bs != 0 {
		return nil, errors.New("wrong crypted size")
	}

	for len(crypted) > 0 {
		block.Decrypt(dst, crypted[:bs])
		crypted = crypted[bs:]
		dst = dst[bs:]
	}

	return out, nil
}

func EncryptMapConnectionInfo(mp map[string]any) (string, string, error) {
	EncodedString, msg := "", ""
	var err error
	shared.Block{
		Try: func() {
			json, err := json.Marshal(mp)
			shared.CheckErr(err)

			EncodedString, msg, err = EncryptMessage(string(json), os.Getenv("CK"))
			shared.CheckErr(err)
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return EncodedString, msg, err
}

func DecryptMapConnectionInfo(str string) map[string]any {
	var mp map[string]interface{}
	shared.Block{
		Try: func() {
			decodeString, msg, err := DecryptMessage(str, os.Getenv("CK"))
			log.Println(msg)
			shared.CheckErr(err)

			json.Unmarshal([]byte(decodeString), &mp)
		},
		Catch: func(e shared.Exception) {
			mp = make(map[string]any)
		},
	}.Do()

	return mp
}

func GenerateJwt(mp map[string]any, key string, duration time.Duration) (string, error) {
	mySigningKey := []byte(key)
	var tokenString string
	var err error

	shared.Block{
		Try: func() {
			token := jwt.New(jwt.SigningMethodHS256)
			claims := token.Claims.(jwt.MapClaims)

			for key, val := range mp {
				claims[key] = val
			}

			claims["exp"] = jwt.NewNumericDate(time.Now().Add(duration))

			tokenString, err = token.SignedString(mySigningKey)
			shared.CheckErr(err)
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return tokenString, err
}

func JwtTokenToMap(tokenString, key string) (map[string]any, error) {
	mapRes := make(map[string]any)
	var err error
	shared.Block{
		Try: func() {
			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(key), nil
			})

			shared.CheckErr(err)

			for key, val := range claims {
				mapRes[key] = val
			}
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return mapRes, err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
