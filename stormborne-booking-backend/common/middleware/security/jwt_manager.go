package security

import (
	"errors"
	"skyfox/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"userId"`
	Role   Role  `json:"role"`
	jwt.RegisteredClaims
}

type JwtManager struct {
	secret      []byte
	tokenExpiry time.Duration
	issuer      string
}

func NewJwtManager(tokenConf config.TokenConfig) *JwtManager {
	return &JwtManager{
		secret:      []byte(tokenConf.SECRET),
		tokenExpiry: time.Second * time.Duration(tokenConf.TTL),
		issuer:      "skyfox-api",
	}
}

func (j *JwtManager) GenerateToken(userID int64, role Role) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   string(rune(userID)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

func (j *JwtManager) ParseToken(tokenStr string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}

			return j.secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.Issuer != j.issuer {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}