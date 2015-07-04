package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"time"
)

type User struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Device    int    `json:"device"`
	DevId     string `json:"dev_id"`
}

type TokenInfo struct {
	Id             string
	ExpirationDate time.Time
}

type AuthToken struct {
	Info string // Contains TokenInfo in base 64 encoded json
	HMAC string // Base 64 encoded hmac
}

func NewAuthToken(user User, expiration_date time.Time, secret string) *AuthToken {
	at := &AuthToken{
		Info: NewTokenInfo(user, expiration_date).ToBase64(),
	}
	at.HMAC = ComputeHmac256(at.Info, secret)
	return at
}

func NewTokenInfo(user User, expiration_date time.Time) *TokenInfo {
	return &TokenInfo{
		Id:             user.Id,
		ExpirationDate: expiration_date,
	}
}

func (at *AuthToken) verify(secret string) bool {
	if ComputeHmac256(at.Info, secret) == at.HMAC {
		return true
	} else {
		return false
	}
}

func (at *AuthToken) GetTokenInfo(secret string) (*TokenInfo, error) {
	/* If the token is not valid, stop now. */
	if !at.verify(secret) {
		return nil, errors.New("The token is not valid.")
	}

	/* Convert from base64. */
	jsonString, err := base64.StdEncoding.DecodeString(at.Info)
	if err != nil {
		log.Fatal("Failed to decode base64 string: ", err)
	}
	/* Unmarshal json object. */
	var ti TokenInfo
	err = json.Unmarshal(jsonString, &ti)
	if err != nil {
		log.Fatal("Failed to decode TokenInfo: ", err)
	}

	//	/* Check if the token is expired. */
	//	if time.Now().Unix() > ti.ExpirationDate.Unix() {
	//		return nil, errors.New("The token is expired.")
	//	} else {
	//		return &ti, nil
	//	}
	return &ti, nil
}

func (ti *TokenInfo) ToBase64() string {
	bytes, err := json.Marshal(ti)
	if err != nil {
		log.Panic("Failed to marshal TokenInfo.")
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func ComputeHmac256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
