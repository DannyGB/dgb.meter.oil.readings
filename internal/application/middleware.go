package application

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

type Middleware struct {
	response *Response
}

func (middleware *Middleware) Options(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			middleware.response.Write(w, 200, nil)
			return
		}

		next(w, r)
	}
}

func (middleware *Middleware) Authorize(next http.HandlerFunc, accessClaim string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := middleware.verifyToken(r, accessClaim)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (middleware *Middleware) extractToken(r *http.Request) string {
	header := r.Header.Get("authorization")
	return strings.Trim(strings.ReplaceAll(header, "Bearer", ""), " ")
}

func (middleware *Middleware) verifyToken(r *http.Request, accessClaim string) (*jwt.Token, error) {
	tokenString := middleware.extractToken(r)
	keySet, err := jwk.Fetch(r.Context(), "https://login.microsoftonline.com/common/discovery/v2.0/keys")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwa.RS256.String() {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		keys, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("key %v not found", kid)
		}

		publickey := &rsa.PublicKey{}
		err = keys.Raw(publickey)
		if err != nil {
			return nil, fmt.Errorf("could not parse pubkey")
		}

		return publickey, nil
	})

	if err != nil {
		return nil, err
	}

	hasAccess, ok := middleware.checkClaim(token, accessClaim)
	if !ok || !hasAccess || !token.Valid {
		return nil, fmt.Errorf("Missing claim")
	}

	return token, nil
}

func (middleware *Middleware) checkClaim(token *jwt.Token, accessClaim string) (bool, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims["scp"] == accessClaim, ok
}

func NewMiddleware(response *Response) *Middleware {
	return &Middleware{
		response,
	}
}
