package auth

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenExpire  = 15 * time.Minute
	RefreshTokenExpire = 7 * 24 * time.Hour
	SigningKey         = "verystrongsecretkey" // В реальном проекте просто из .env переменной подгружать
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var (
	Tokens     = make(map[string]string) // key = refresh, value = access
	lockTokens = &sync.RWMutex{}
)

func GenerateTokens(userID string) (accessToken string, refreshToken string, err error) {
	accessClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(SigningKey))
	if err != nil {
		return "", "", fmt.Errorf("error creating access token: %w", err)
	}

	refreshClaims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpire)),
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SigningKey))
	if err != nil {
		return "", "", fmt.Errorf("error creating refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SigningKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func ValidateTokenPair(accessToken, refreshToken string) (string, error) {

	accessClaims, err := ParseToken(accessToken)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		log.Printf("invalif access token %v", err)
		return "", fmt.Errorf("invalid access token: %v", err)
	}

	refreshClaims, err := ParseToken(refreshToken)
	if err != nil {
		log.Printf("invalif refresh token %v", err)
		return "", fmt.Errorf("invalid refresh token: %v", err)
	}

	if accessClaims.UserID != refreshClaims.UserID {
		log.Printf("tokens belong to different users")
		return "", errors.New("tokens belong to different users")
	}

	return accessClaims.UserID, nil
}

func CheckTokensExists(access, refresh string) error {
	value, exists := getTokens(access)

	if !exists {
		return fmt.Errorf("not exists token")
	}

	if value != refresh {
		return fmt.Errorf("refresh token not validate access")
	}

	return nil
}

func AddTokensToBd(access, refresh string) error {

	_, exists := getTokens(access)

	if exists {
		return fmt.Errorf("cannot add tokens: already exists")
	}

	setTokens(access, refresh)

	return nil
}

func DeleteTokens(oldAccess, oldRefresh string) error {
	if err := CheckTokensExists(oldAccess, oldRefresh); err != nil {
		return fmt.Errorf("cannot delete tokens: old tokens are not correct")
	}

	lockTokens.Lock()
	delete(Tokens, oldAccess)
	lockTokens.Unlock()

	return nil
}

func getTokens(access string) (string, bool) {
	lockTokens.Lock()
	value, exists := Tokens[access]
	lockTokens.Unlock()
	return value, exists
}

func setTokens(access, refresh string) {
	lockTokens.Lock()
	Tokens[access] = refresh
	lockTokens.Unlock()
}
