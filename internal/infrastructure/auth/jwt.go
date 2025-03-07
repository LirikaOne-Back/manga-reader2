package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"manga-reader2/internal/domain/entity"
)

// JWTService предоставляет функции для работы с JWT
type JWTService struct {
	accessSecret   string
	refreshSecret  string
	accessExpires  time.Duration
	refreshExpires time.Duration
}

// Claims содержит данные, которые будут сохранены в токене
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTService создает новый экземпляр JWTService
func NewJWTService(
	accessSecret string,
	refreshSecret string,
	accessExpHours int,
	refreshExpDays int,
) *JWTService {
	return &JWTService{
		accessSecret:   accessSecret,
		refreshSecret:  refreshSecret,
		accessExpires:  time.Duration(accessExpHours) * time.Hour,
		refreshExpires: time.Duration(refreshExpDays) * 24 * time.Hour,
	}
}

// GenerateTokenPair создает новую пару токенов: access и refresh
func (s *JWTService) GenerateTokenPair(user *entity.User) (*entity.TokenPair, error) {
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания refresh token: %w", err)
	}

	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateAccessToken создает новый access token для пользователя
func (s *JWTService) GenerateAccessToken(user *entity.User) (string, error) {
	expirationTime := time.Now().Add(s.accessExpires)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken создает новый refresh token для пользователя
func (s *JWTService) GenerateRefreshToken(userID int64) (string, error) {
	expirationTime := time.Now().Add(s.refreshExpires)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateAccessToken проверяет валидность access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.accessSecret)
}

// ValidateRefreshToken проверяет валидность refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.refreshSecret)
}

// validateToken проверяет валидность JWT токена
func (s *JWTService) validateToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("токен истек")
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("недействительный токен")
	}

	return claims, nil
}

// RefreshTokens обновляет пару токенов, используя refresh token
func (s *JWTService) RefreshTokens(refreshToken string, user *entity.User) (*entity.TokenPair, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.UserID != user.ID {
		return nil, errors.New("user_id не соответствует refresh token")
	}

	return s.GenerateTokenPair(user)
}
