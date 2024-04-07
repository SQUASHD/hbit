package http

import (
	"log"
	"net/http"
	"time"

	"github.com/SQUASHD/hbit"
	"github.com/SQUASHD/hbit/auth"
	"github.com/SQUASHD/hbit/config"
)

// Middleware type definition
type Middleware func(http.Handler) http.Handler

// ChainMiddleware chains multiple middleware functions together
func ChainMiddleware(handler http.Handler, middleware ...Middleware) http.Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

// customResponseWriter is a wrapper around an http.ResponseWriter that keeps track of the response status code
// deprecated
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// internalAuthMiddleware is a middleware that check if the 'X-Internal-Request'
// header is set to 'true'. This is by default set to 'false' in the SetInternalHeaderMiddleware
// which is wrapped around the gateway router
func internalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := getInternalHeader(r)
		if err != nil {
			Error(w, r, err)
			return
		}
		next(w, r)
	}
}

// SetInternalHeaderMiddleware sets the 'X-Internal-Request' header to 'false'
// This ensures that we can proctect some routes from being accessed via the API Gateway
// even by authenticated users
func SetInternalHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Internal-Request", "false")
		next.ServeHTTP(w, r)
	})
}

// getInternalHeader is a helper function for InternalAuthMiddleware
func getInternalHeader(r *http.Request) error {
	internal := r.Header.Get("X-Internal-Request")
	if internal == "" || internal == "false" {
		log.Printf("Unauthorized request: %s", r.URL.Path)
		return &hbit.Error{Code: hbit.EUNAUTHORIZED, Message: "Unauthorized"}
	}
	return nil
}

// setInternalHeader is a helper function for SetInternalHeaderMiddleware
func setInternalHeader(r *http.Request) {
	r.Header.Set("X-Internal-Request", "true")
}

// NewCustomResponseWriter was middleware used to note the status code of the response
// Howver, it causes issues when using an API Gateway
func NewCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	return &customResponseWriter{w, http.StatusOK}
}

// WriteHeader implements the http.ResponseWriter interface
func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// LoggerMiddleware logs the request method, URL path, and duration of the request
// deprecated
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		crw := NewCustomResponseWriter(w)
		next.ServeHTTP(crw, r)
		log.Println(r.Method, crw.statusCode, r.URL.Path, time.Since(start))
	})
}

// AuthedHandler is a type definition for a handler that requires authentication
// Most routes go through this middleware
type AuthedHandler func(w http.ResponseWriter, r *http.Request, userId string)

// AuthChainMiddleware is a higher order function that returns a middleware function that authenticates users
func AuthChainMiddleware(userIdGetter func(r *http.Request) (string, error)) func(next AuthedHandler) http.HandlerFunc {
	return func(next AuthedHandler) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userId, err := userIdGetter(r)
			if err != nil {
				Error(w, r, err)
				return
			}
			next(w, r, userId)
		})
	}
}

func AuthMiddleware(next AuthedHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := GetUserIdFromHeader(r)
		if err != nil {
			Error(w, r, err)
			return
		}
		next(w, r, userId)
	})
}

// GetUserIdFromHeader is a helper function that extracts the X-User-Id header from a request
func GetUserIdFromHeader(r *http.Request) (string, error) {
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		log.Printf("Unauthorized request: %s", r.URL.Path)
		return "", &hbit.Error{Code: hbit.EUNAUTHORIZED, Message: "Missing user id header"}
	}
	return userId, nil
}

// CORSMiddlware... the name's on the tin
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		next.ServeHTTP(w, r)
	})
}

// JwtAuthRouterMiddleware is middleware that authenticates user before forwarding request to the router
// Refactor from being a method of authHandler
func JwtAuthRouterMiddleware(svc auth.JwtAuth, jwtConf config.JwtOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userId, err := authenticateUser(w, r, svc, jwtConf)
			if err != nil {
				Error(w, r, err)
				return
			}
			setUserIdInRequestHeader(r, userId)
			next.ServeHTTP(w, r)
		})
	}
}

// setUserIdInRequestHeader is a helper function to JwtAuthMiddleware
// since user authentication is done at the API gateway via cookies, the user id is set in the request header
func setUserIdInRequestHeader(r *http.Request, userId string) {
	r.Header.Set("X-User-Id", userId)
}

// authenticateUser is a helper function to JwtAuthMiddleware
func authenticateUser(w http.ResponseWriter, r *http.Request, svc auth.JwtAuth, jwtConf config.JwtOptions) (string, error) {
	refreshToken := getRefreshTokenFromCookie(r)
	accessToken := getAccessTokenFromCookie(r)
	// If refresh token is missing, clear all tokens and return error
	if refreshToken == "" {
		clearTokensFromCookie(w)
		return "", &hbit.Error{Code: hbit.EUNAUTHORIZED, Message: "Missing tokens"}
	}
	// If access token is missing, refresh token
	if accessToken == "" {
		accessToken, userId, err := svc.RefreshToken(r.Context(), refreshToken)
		if err != nil {
			return "", err
		}
		setAccessCookie(w, accessToken, jwtConf.AccessDuration)
		return userId, nil
	}
	// If both tokens are present, authenticate user
	userId, err := svc.AuthenticateUser(r.Context(), accessToken)
	if err != nil {
		return "", err
	}
	return userId, nil
}
