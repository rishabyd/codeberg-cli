package codeberg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
	"golang.org/x/oauth2"
)

var oauthCfg = &oauth2.Config{
	ClientID: constants.OAuthClientID,
	Endpoint: oauth2.Endpoint{
		AuthURL:  constants.OAuthAuthorizeURL,
		TokenURL: constants.OAuthTokenURL,
	},
	RedirectURL: constants.OAuthRedirectURI,
}

var restyClient = resty.New().
	SetBaseURL(constants.CodebergAPIBaseURL).
	SetTimeout(30 * time.Second).
	SetRetryCount(3).
	SetRetryWaitTime(1 * time.Second).
	SetRetryMaxWaitTime(10 * time.Second).
	SetHeader("Accept", "application/json").
	AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return true
		}
		code := r.StatusCode()
		return code >= 500 || code == 429
	}).
	OnAfterResponse(func(_ *resty.Client, r *resty.Response) error {
		if r.StatusCode() == http.StatusTooManyRequests {
			return fmt.Errorf("rate limited by Codeberg — wait and try again")
		}
		return nil
	})

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type Repo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
	CloneURL    string `json:"clone_url"`
	SSHURL      string `json:"ssh_url"`
	HTMLURL     string `json:"html_url"`
}

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	AutoInit    bool   `json:"auto_init"`
	DefaultRef  string `json:"default_branch,omitempty"`
}

type MigrateRepoRequest struct {
	Service      string `json:"service"`
	CloneAddr    string `json:"clone_addr"`
	RepoName     string `json:"repo_name"`
	Private      bool   `json:"private"`
	Issues       bool   `json:"issues"`
	PullRequests bool   `json:"pull_requests"`
	Wiki         bool   `json:"wiki"`
	Labels       bool   `json:"labels"`
	Milestones   bool   `json:"milestones"`
	Releases     bool   `json:"releases"`
}

type apiError struct {
	Message string `json:"message"`
}

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "Authentication failed"
	}
	return e.Message
}

type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	if e == nil || e.Err == nil {
		return "Network error"
	}
	return e.Err.Error()
}

func (e *NetworkError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsAuthError(err error) bool {
	var target *AuthError
	return errors.As(err, &target)
}

func IsNetworkError(err error) bool {
	var target *NetworkError
	return errors.As(err, &target)
}

// ---- OAuth endpoints ----

func GenerateVerifier() string {
	return oauth2.GenerateVerifier()
}

func AuthorizationURL(state, verifier string) string {
	return oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
}

func ExchangeCode(ctx context.Context, code, verifier string) (*TokenResponse, error) {
	tok, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, err
	}
	return tokenToResponse(tok), nil
}

func refreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	tok := &oauth2.Token{RefreshToken: refreshToken}
	ts := oauthCfg.TokenSource(ctx, tok)
	newTok, err := ts.Token()
	if err != nil {
		return nil, err
	}
	if newTok.AccessToken == "" {
		return nil, fmt.Errorf("refresh failed: no access token")
	}
	return tokenToResponse(newTok), nil
}

func tokenToResponse(tok *oauth2.Token) *TokenResponse {
	return &TokenResponse{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		ExpiresIn:    int64(time.Until(tok.Expiry).Seconds()),
	}
}

// ---- Auth helpers ----

func GetCurrentUserByToken(ctx context.Context, accessToken string) (*User, int, error) {
	var u User
	resp, err := restyClient.R().
		SetContext(ctx).
		SetAuthToken(accessToken).
		SetResult(&u).
		Get(constants.CodebergAPIBaseURL + "/user")

	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode() >= 300 {
		return nil, resp.StatusCode(), nil
	}
	return &u, resp.StatusCode(), nil
}

// ---- API methods ----

func GetCurrentUser(ctx context.Context, cfg *config.AuthConfig) (*User, error) {
	var u User
	_, err := doAuthenticatedResty(ctx, cfg, resty.MethodGet, func(r *resty.Request) {
		r.SetResult(&u)
	}, "/user")
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func FetchUserRepos(ctx context.Context, cfg *config.AuthConfig, limit int) ([]Repo, error) {
	if limit <= 0 {
		limit = 30
	}
	var out []Repo
	_, err := doAuthenticatedResty(ctx, cfg, resty.MethodGet, func(r *resty.Request) {
		r.SetResult(&out).
			SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}, "/user/repos")
	if err != nil {
		return nil, err
	}
	return out, nil
}

func CreateRepo(ctx context.Context, cfg *config.AuthConfig, payload CreateRepoRequest) (*Repo, error) {
	var out Repo
	_, err := doAuthenticatedResty(ctx, cfg, resty.MethodPost, func(r *resty.Request) {
		r.SetBody(payload).
			SetResult(&out)
	}, "/user/repos")
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func MigrateRepo(ctx context.Context, cfg *config.AuthConfig, payload MigrateRepoRequest) (*Repo, error) {
	var out Repo
	_, err := doAuthenticatedResty(ctx, cfg, resty.MethodPost, func(r *resty.Request) {
		r.SetBody(payload).
			SetResult(&out)
	}, "/repos/migrate")
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ---- internal ----

func doAuthenticatedResty(ctx context.Context, cfg *config.AuthConfig, method string, setup func(r *resty.Request), path string) (*resty.Response, error) {
	if cfg == nil || strings.TrimSpace(cfg.AccessToken) == "" {
		return nil, &AuthError{Message: "Not logged in"}
	}

	if time.Now().After(cfg.Expiry) && strings.TrimSpace(cfg.RefreshToken) != "" {
		if err := refreshAndSave(ctx, cfg); err != nil {
			return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
		}
	}

	req := restyClient.R().
		SetContext(ctx).
		SetAuthToken(cfg.AccessToken)

	setup(req)

	resp, err := req.Execute(method, path)
	if err != nil {
		return nil, wrapNetworkError(err)
	}

	if resp.StatusCode() != http.StatusUnauthorized && resp.StatusCode() != http.StatusForbidden {
		if resp.StatusCode() >= 400 {
			msg := extractAPIError(resp.Body(), resp.StatusCode())
			return nil, fmt.Errorf("%s", msg)
		}
		return resp, nil
	}

	if strings.TrimSpace(cfg.RefreshToken) == "" {
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}

	if err := refreshAndSave(ctx, cfg); err != nil {
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}

	req2 := restyClient.R().
		SetContext(ctx).
		SetAuthToken(cfg.AccessToken)

	setup(req2)

	resp2, err := req2.Execute(method, path)
	if err != nil {
		return nil, wrapNetworkError(err)
	}
	if resp2.StatusCode() == http.StatusUnauthorized || resp2.StatusCode() == http.StatusForbidden {
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}
	if resp2.StatusCode() >= 400 {
		msg := extractAPIError(resp2.Body(), resp2.StatusCode())
		return nil, fmt.Errorf("%s", msg)
	}
	return resp2, nil
}

func refreshAndSave(ctx context.Context, cfg *config.AuthConfig) error {
	refreshed, err := refreshToken(ctx, cfg.RefreshToken)
	if err != nil {
		return err
	}
	cfg.AccessToken = refreshed.AccessToken
	if strings.TrimSpace(refreshed.RefreshToken) != "" {
		cfg.RefreshToken = refreshed.RefreshToken
	}
	cfg.Expiry = time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second)
	return config.Save(*cfg)
}

func extractAPIError(body []byte, code int) string {
	if len(body) == 0 {
		return fmt.Sprintf("HTTP %d", code)
	}
	var apiErr apiError
	if err := json.Unmarshal(body, &apiErr); err == nil && strings.TrimSpace(apiErr.Message) != "" {
		return apiErr.Message
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Sprintf("HTTP %d", code)
	}
	return msg
}

func wrapNetworkError(err error) error {
	if err == nil {
		return nil
	}
	return &NetworkError{Err: err}
}
