package token

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
	"time"
)

type (
	//TokenImpl struct for keeping info about token
	tokenImpl struct {
		Id        string
		Username  string
		UserId    int64
		Role      string
		LoginDate time.Time
		Expires   time.Time
	}

	CustomizedClaims struct {
		jwt.StandardClaims
		Role string `json:"role"`
	}
)

const admin = "admin"
const user = "user"

func (token *tokenImpl) GetId() string {
	return token.Id
}

func (token *tokenImpl) GetUserId() int64 {
	return token.UserId
}

func (token *tokenImpl) GetLoginDate() time.Time {
	return token.LoginDate
}

func (token *tokenImpl) GetExpirationTime() time.Time {
	return token.Expires
}

func (token *tokenImpl) GetRole() string {
	return token.Role
}

func (token *tokenImpl) GetUsername() string {
	return token.Username
}

//func (token *tokenImpl) GetLoginDateFormatted() string {
//	tm := time.Unix(int64(token.LoginDate), 0)
//	return tm.Format("2006-01-02 15:04:05")
//}

//GenerateTokenID creates new random token id
func GenerateTokenID() (string, error) {
	size := 30
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Error("failed to create token", err)
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}
