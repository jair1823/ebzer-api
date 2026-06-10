package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Role    Role   `json:"role"`
	Type    string `json:"type"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) Generate(user *User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Subject: strconv.Itoa(user.ID),
		Email:   user.Email,
		Role:    user.Role,
		Type:    tokenType,
		Issued:  now.Unix(),
		Expires: now.Add(ttl).Unix(),
	}
	return s.sign(claims)
}

func (s *JWTService) Validate(token string, expectedType string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := s.signature(signingInput)
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actualSignature, expectedSignature) {
		return nil, errors.New("invalid token")
	}

	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, errors.New("invalid token")
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return nil, errors.New("invalid token")
	}

	var claims Claims
	if err := decodeSegment(parts[1], &claims); err != nil {
		return nil, errors.New("invalid token")
	}
	if claims.Type != expectedType {
		return nil, errors.New("invalid token type")
	}
	if claims.Expires <= time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	if claims.Subject == "" || claims.Email == "" || !claims.Role.IsValid() {
		return nil, errors.New("invalid token claims")
	}
	return &claims, nil
}

func (s *JWTService) sign(claims Claims) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
		base64.RawURLEncoding.EncodeToString(claimsBytes)
	signature := base64.RawURLEncoding.EncodeToString(s.signature(signingInput))
	return fmt.Sprintf("%s.%s", signingInput, signature), nil
}

func (s *JWTService) signature(signingInput string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}

func decodeSegment(segment string, destination any) error {
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, destination)
}
