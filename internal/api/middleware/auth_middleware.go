package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
	"redo.ai/internal/utils"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

// CustomClaims defines any custom claims you want to use.
type CustomClaims struct {
	Scope string `json:"scope"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken sets up Auth0 JWT validation middleware.
func EnsureValidToken() func(http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the JWT validator: %v", err)
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("JWT validation error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

// WithUser populates the user ID into the request context.
func WithUser(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*jwt.Token)
			if !ok || token == nil {
				utils.WriteJSONError(w, http.StatusUnauthorized, "missing token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				utils.WriteJSONError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				utils.WriteJSONError(w, http.StatusUnauthorized, "missing sub claim")
				return
			}

			var userID string
			err := db.QueryRow(`SELECT id FROM users WHERE email = $1`, sub).Scan(&userID)
			if err == sql.ErrNoRows {
				utils.WriteJSONError(w, http.StatusUnauthorized, "user not found")
				return
			} else if err != nil {
				utils.WriteJSONError(w, http.StatusInternalServerError, "database error")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext retrieves the user ID from the request context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}
