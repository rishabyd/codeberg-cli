package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/rishabyd/codeberg-cli/internal/constants"
)

func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

func AuthorizationURL(challenge string) string {
	v := url.Values{}
	v.Set("client_id", constants.OAuthClientID)
	v.Set("redirect_uri", constants.OAuthRedirectURI)
	v.Set("response_type", "code")
	v.Set("code_challenge_method", "S256")
	v.Set("code_challenge", challenge)
	return fmt.Sprintf("%s?%s", constants.OAuthAuthorizeURL, v.Encode())
}
