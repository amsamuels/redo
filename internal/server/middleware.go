// internal/middleware/middleware.go
package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	companyIDKey ctxKey = "company_id"
)

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

// WithCompany extracts the sub from the token and ensures a company exists.
func WithUser(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*jwt.Token)
			if !ok || token == nil {
				http.Error(w, "Token not found in context", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				http.Error(w, "Missing subject claim", http.StatusUnauthorized)
				return
			}

			var companyID string
			err := db.QueryRow(`SELECT id FROM companies WHERE email = $1`, sub).Scan(&companyID)
			if err == sql.ErrNoRows {
				err = db.QueryRow(`
					INSERT INTO companies (id, name, email)
					VALUES (gen_random_uuid(), $1, $1)
					RETURNING id
				`, sub).Scan(&companyID)
				if err != nil {
					http.Error(w, "Unable to create company", http.StatusInternalServerError)
					return
				}
			} else if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), companyIDKey, companyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CompanyIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(companyIDKey).(string)
	return id, ok
}

func (c CustomClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for _, s := range result {
		if s == expectedScope {
			return true
		}
	}
	return false
}

func Wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h(w, r)
	}
}
