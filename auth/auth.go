package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/piskovoy-dmitrij/MoC-pulse-backend/log"
)

var secret string = "shjgfshfkjgskdfjgksfghks"

type User struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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

func NewAuthToken(user User, expiration_date time.Time) *AuthToken {
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

func (at *AuthToken) GetTokenInfo() (*TokenInfo, error) {
	funcPrefix := "Getting token info"
	log.Debug.Printf("%s: start\n", funcPrefix)
	defer log.Debug.Printf("%s: end\n", funcPrefix)

	/* If the token is not valid, stop now. */
	if !at.verify(secret) {
		log.Error.Printf("%s: token is not valid\n", funcPrefix)
		return nil, errors.New("The token is not valid.")
	}

	/* Convert from base64. */
	jsonString, err := base64.StdEncoding.DecodeString(at.Info)
	if err != nil {
		log.Error.Printf("%s: decoding base64 string failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	/* Unmarshal json object. */
	var ti TokenInfo
	log.Debug.Printf("%s: unmarshaling token info...\n", funcPrefix)
	err = json.Unmarshal(jsonString, &ti)
	if err != nil {
		log.Error.Printf("%s: unmarshaling token info failed: %s\n", funcPrefix, err.Error())
		return nil, err
	}

	//	/* Check if the token is expired. */
	//	if time.Now().Unix() > ti.ExpirationDate.Unix() {
	//		log.Error.Printf("%s: token is expired\n", funcPrefix)
	//		return nil, errors.New("The token is expired.")
	//	}

	return &ti, nil
}

func (ti *TokenInfo) ToBase64() string {
	bytes, err := json.Marshal(ti)
	if err != nil {
		log.Error.Printf("Marshaling token info failed: %s\n", err.Error())
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func ComputeHmac256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
