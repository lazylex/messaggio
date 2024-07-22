package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lazylex/messaggio/internal/helpers/constants/prefixes"
	"github.com/lazylex/messaggio/internal/helpers/constants/various"
	"log/slog"
	"net/http"
	"strings"
)

const (
	requestHeaderPrefix = "Bearer "
	header              = "Authorization"
)

var (
	validMethods = []string{"HS256"}

	ErrNoExpirationClaims = errors.New("expiration date of the token is absent")
)

type MiddlewareJWT struct {
	secret []byte // Секретный ключ, которым должны быть подписаны валидные токены
}

// New конструктор прослойки для проверки JSON Web Token
func New(secret []byte) *MiddlewareJWT {
	return &MiddlewareJWT{secret: secret}
}

// CheckJWT проверяет JWT токен в запросе. В случае, если токен не валидный, функция прекращает дальнейшую обработку
// запроса сервисом. Ошибка заносится в лог, отправителю возвращается ответ с кодом http.StatusUnauthorized. Проверяется
// только срок годности токена.
func (m *MiddlewareJWT) CheckJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		uri := r.URL.RequestURI()
		if strings.HasPrefix(uri, prefixes.PPROFPrefix) || strings.HasPrefix(uri, "/favicon.ico") {
			next.ServeHTTP(rw, r)
			return
		}

		var notParsedToken string
		log := slog.Default().With(various.Origin, "adapters.http.middlewares.jwt.CheckJWT")

		if len(r.Header.Get(header)) > len(requestHeaderPrefix) {
			notParsedToken = r.Header.Get(header)[len(requestHeaderPrefix):]
		} else {
			log.Error("no JWT token find")
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(
			notParsedToken,
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return m.secret, nil
			},
			jwt.WithValidMethods(validMethods),
		)

		if !token.Valid {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				log.Warn("not a JWT token in request")
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				log.Warn("invalid JWT token signature")
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				log.Warn("token expired")
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				log.Warn("token not valid yet")
			} else {
				log.Warn("couldn't handle this token:", err)
			}

			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if _, ok = claims["exp"]; ok {
				next.ServeHTTP(rw, r)
				return
			}
		}

		log.Warn(ErrNoExpirationClaims.Error())
		rw.WriteHeader(http.StatusUnauthorized)
		return
	})
}
