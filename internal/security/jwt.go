package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtIssuer   = "FinanceMaster"
	jwtAudience = "finance-master-mobile"
	jwtSkew     = 30 * time.Second

	primary   []byte
	fallbacks [][]byte
)

func InitKeyRing(primaryKey []byte, fallbackKeys [][]byte, iss, aud string, skew time.Duration) {
	primary = primaryKey
	fallbacks = fallbackKeys
	if iss != "" {
		jwtIssuer = iss
	}
	if aud != "" {
		jwtAudience = aud
	}
	if skew > 0 {
		jwtSkew = skew
	}
}

func SignAccess(secret []byte, uid int64, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": uid, "iss": jwtIssuer, "aud": jwtAudience, "typ": "access",
		"iat": now.Unix(), "nbf": now.Unix(), "exp": now.Add(ttl).Unix(), "kid": 0,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}
func SignRefresh(secret []byte, uid int64, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": uid, "iss": jwtIssuer, "aud": jwtAudience, "typ": "refresh",
		"iat": now.Unix(), "nbf": now.Unix(), "exp": now.Add(ttl).Unix(), "kid": 0,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

var ErrInvalidJWT = errors.New("invalid_jwt")

func Parse(_ []byte, token string) (jwt.MapClaims, error) {
	tryKeys := make([][]byte, 0, 1+len(fallbacks))
	if len(primary) > 0 {
		tryKeys = append(tryKeys, primary)
	}
	tryKeys = append(tryKeys, fallbacks...)

	var _ error
	for _, k := range tryKeys {
		mc := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, mc, func(t *jwt.Token) (interface{}, error) {
			return k, nil
		}, jwt.WithLeeway(jwtSkew))
		if err != nil {
			_ = err
			continue
		}
		// iss
		if iss, _ := mc["iss"].(string); iss != jwtIssuer {
			_ = ErrInvalidJWT
			continue
		}
		// aud
		okAud := false
		switch v := mc["aud"].(type) {
		case string:
			okAud = v == jwtAudience
		case []any:
			for _, a := range v {
				if s, _ := a.(string); s == jwtAudience {
					okAud = true
					break
				}
			}
		}
		if !okAud {
			_ = ErrInvalidJWT
			continue
		}

		now := time.Now()
		if exp, ok := mc["exp"].(float64); ok && now.After(time.Unix(int64(exp), 0).Add(jwtSkew)) {
			_ = ErrInvalidJWT
			continue
		}
		if nbf, ok := mc["nbf"].(float64); ok && now.Before(time.Unix(int64(nbf), 0).Add(-jwtSkew)) {
			_ = ErrInvalidJWT
			continue
		}
		return mc, nil
	}
	return nil, ErrInvalidJWT
}

func ActiveSigningKey() []byte { return primary }
