package token

import (
	"SB/catb406.github.io/service/config"
	"SB/catb406.github.io/service/repository/persistence"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type (
	TokenManager interface {
		GenerateNewToken(userId int64, role string, username string) (accessToken persistence.Token, refreshToken persistence.Token, err error)
		Add(token persistence.Token) error
		Get(id string) (persistence.Token, error)
		//GetAll() ([]Token, error)
		Remove(id string) error
		Refresh(tokenId string) (accessTkn persistence.Token, refreshTkn persistence.Token, err error)
		//RemoveAllTokens() error
		//WatchExpired() (error)
		//KeepAlive(id TokenID) error
	}

	TokenManagerImpl struct {
		db persistence.Persistent
	}
)

func NewTokenManager(persistent persistence.Persistent) TokenManager {
	return &TokenManagerImpl{
		db: persistent,
	}
}

func (mgr *TokenManagerImpl) GenerateNewToken(userId int64, role string, username string) (accessTkn persistence.Token, refreshTkn persistence.Token, err error) {
	//claims := CustomizedClaims{
	//	StandardClaims: jwt.StandardClaims{
	//		Id:        fmt.Sprintf("%d", userId),
	//		ExpiresAt: time.Now().UTC().Add(config.AccessTokenExpiration).Unix(),
	//	},
	//	Role: role,
	//}

	claims := jwt.StandardClaims{
		Id:        fmt.Sprintf("%d", userId),
		ExpiresAt: time.Now().UTC().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	accessTokenId, err := token.SignedString([]byte(config.Secret))
	refreshTokenId, err := GenerateTokenID()
	if err != nil {
		return nil, nil, err
	}
	refreshTkn = &tokenImpl{
		Username:  username,
		Id:        refreshTokenId,
		UserId:    userId,
		Role:      role,
		LoginDate: time.Now().UTC(),
		Expires:   time.Now().UTC().Add(config.RefreshTokenExpiration),
	}
	err = mgr.Add(refreshTkn)
	if err != nil {
		return nil, nil, err
	}
	accessTkn = &tokenImpl{
		Username:  username,
		Role:      role,
		Id:        accessTokenId,
		Expires:   time.Unix(claims.ExpiresAt, 0),
		LoginDate: time.Now().UTC(),
		UserId:    userId,
	}
	return
}

func (mgr *TokenManagerImpl) Add(token persistence.Token) error {
	added := mgr.db.AddSession(token)
	if !added {
		return errors.New("failed to add token")
	}
	return nil
}
func (mgr *TokenManagerImpl) Get(tokenId string) (persistence.Token, error) {
	return mgr.db.GetSession(tokenId)
}

func (mgr *TokenManagerImpl) Refresh(tokenId string) (accessTkn persistence.Token, refreshTkn persistence.Token, err error) {
	tkn, err := mgr.db.GetSession(tokenId)
	if err != nil {
		return nil, nil, err
	}
	err = mgr.db.RemoveSession(tokenId)
	if err != nil {
		return nil, nil, err
	}
	if tkn.GetExpirationTime().Unix() < time.Now().UTC().Unix() {
		return nil, nil, errors.New("refresh token expired")
	}
	return mgr.GenerateNewToken(tkn.GetUserId(), tkn.GetRole(), tkn.GetUsername())
}

// TODO: implement all tokens method
//func (mgr *TokenManagerImpl) GetAll() ([]Token, error) {
//
//}

func (mgr *TokenManagerImpl) Remove(id string) error {
	return mgr.db.RemoveSession(id)
}

//func (mgr *TokenManagerImpl) RemoveAllTokens() error{}
//func (mgr *TokenManagerImpl) WatchExpired() (error){}

// TODO: implement expired token monitoring
//func (mgr *TokenManagerImpl) KeepAlive(id TokenID) error{}
