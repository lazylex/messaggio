package http

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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

// NewJWTMiddleware конструктор прослойки для проверки JSON Web Token.
func NewJWTMiddleware(secret []byte) *MiddlewareJWT {
	return &MiddlewareJWT{secret: secret}
}

// CheckJWT проверяет JWT токен в запросе. В случае, если токен не валидный, функция прекращает дальнейшую обработку
// запроса сервисом. Ошибка заносится в лог, отправителю возвращается ответ с кодом http.StatusUnauthorized. Проверяется
// только срок годности токена.
func (m *MiddlewareJWT) CheckJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		if strings.HasPrefix(uri, prefixes.PPROFPrefix) || strings.HasPrefix(uri, "/favicon.ico") {
			c.Next()
			return
		}

		var notParsedToken string
		log := slog.Default().With(various.Origin, "adapters.http.middlewares.jwt.CheckJWT")

		if len(c.GetHeader(header)) > len(requestHeaderPrefix) {
			notParsedToken = c.GetHeader(header)[len(requestHeaderPrefix):]
		} else {
			log.Error("no JWT token find")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no JWT token find"})
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
				c.JSON(http.StatusUnauthorized, gin.H{"error": "not a JWT token in request"})
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				log.Warn("invalid JWT token signature")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid JWT token signature"})
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				log.Warn("token expired")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				log.Warn("token not valid yet")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token not valid yet"})
			} else {
				log.Warn("couldn't handle this token:", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "couldn't handle this token:"})
			}

			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if _, ok = claims["exp"]; ok {
				c.Next()
				return
			}
		}

		log.Warn(ErrNoExpirationClaims.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "expiration date of the token is absent"})
		return
	}
}
