package codeberg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/constants"
)

var (
	apiBaseURL    = constants.CodebergAPIBaseURL
	oauthTokenURL = constants.OAuthTokenURL
)

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

func ExchangeCode(ctx context.Context, code, verifier string) (*TokenResponse, error) {
	v := url.Values{}
	v.Set("client_id", constants.OAuthClientID)
	v.Set("grant_type", "authorization_code")
	v.Set("redirect_uri", constants.OAuthRedirectURI)
	v.Set("code", code)
	v.Set("code_verifier", verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, constants.OAuthTokenURL, bytes.NewBufferString(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var out TokenResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.AccessToken == "" {
		return nil, fmt.Errorf("token exchange failed: no access token")
	}
	return &out, nil
}

func RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	v := url.Values{}
	v.Set("client_id", constants.OAuthClientID)
	v.Set("grant_type", "refresh_token")
	v.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, bytes.NewBufferString(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("refresh failed with status %d", res.StatusCode)
	}

	var out TokenResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.AccessToken == "" {
		return nil, fmt.Errorf("refresh failed: no access token")
	}
	return &out, nil
}

func GetCurrentUserByToken(ctx context.Context, accessToken string) (*User, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/user", nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, res.StatusCode, nil
	}

	var u User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, res.StatusCode, err
	}
	return &u, res.StatusCode, nil
}

func GetCurrentUser(ctx context.Context, cfg *config.AuthConfig) (*User, error) {
	res, err := doAuthenticatedRequest(ctx, cfg, http.MethodGet, apiBaseURL+"/user", nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("failed to fetch current user: %s", extractAPIError(res.Body, res.StatusCode))
	}

	var u User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func FetchUserRepos(ctx context.Context, cfg *config.AuthConfig, limit int) ([]Repo, error) {
	if limit <= 0 {
		limit = 30
	}

	res, err := doAuthenticatedRequest(ctx, cfg, http.MethodGet, fmt.Sprintf("%s/user/repos?limit=%d", apiBaseURL, limit), nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("failed to fetch repositories: %s", extractAPIError(res.Body, res.StatusCode))
	}

	var out []Repo
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func CreateRepo(ctx context.Context, cfg *config.AuthConfig, payload CreateRepoRequest) (*Repo, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	res, err := doAuthenticatedRequest(ctx, cfg, http.MethodPost, apiBaseURL+"/user/repos", b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("failed to create repository: %s", extractAPIError(res.Body, res.StatusCode))
	}

	var out Repo
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func MigrateRepo(ctx context.Context, cfg *config.AuthConfig, payload MigrateRepoRequest) (*Repo, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	res, err := doAuthenticatedRequest(ctx, cfg, http.MethodPost, apiBaseURL+"/repos/migrate", b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("migration failed: %s", extractAPIError(res.Body, res.StatusCode))
	}

	var out Repo
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func extractAPIError(r io.Reader, code int) string {
	body, _ := io.ReadAll(io.LimitReader(r, 4096))
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

func doAuthenticatedRequest(ctx context.Context, cfg *config.AuthConfig, method, targetURL string, body []byte, contentType string) (*http.Response, error) {
	if cfg == nil || strings.TrimSpace(cfg.AccessToken) == "" {
		return nil, &AuthError{Message: "Not logged in"}
	}

	res, err := doRequest(ctx, cfg.AccessToken, method, targetURL, body, contentType)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusUnauthorized && res.StatusCode != http.StatusForbidden {
		return res, nil
	}

	res.Body.Close()
	if strings.TrimSpace(cfg.RefreshToken) == "" {
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}

	refreshed, err := RefreshToken(ctx, cfg.RefreshToken)
	if err != nil {
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}

	cfg.AccessToken = refreshed.AccessToken
	if strings.TrimSpace(refreshed.RefreshToken) != "" {
		cfg.RefreshToken = refreshed.RefreshToken
	}
	if saveErr := config.Save(*cfg); saveErr != nil {
		return nil, saveErr
	}

	res, err = doRequest(ctx, cfg.AccessToken, method, targetURL, body, contentType)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		res.Body.Close()
		return nil, &AuthError{Message: "Session expired. Run `cb auth login`"}
	}

	return res, nil
}

func doRequest(ctx context.Context, accessToken, method, targetURL string, body []byte, contentType string) (*http.Response, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if isNetworkFailure(err) {
			return nil, &NetworkError{Err: err}
		}
		return nil, err
	}
	return res, nil
}

func isNetworkFailure(err error) bool {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return true
		}
		inner := urlErr.Err
		if inner != nil {
			return isNetworkFailure(inner)
		}
	}

	return false
}
