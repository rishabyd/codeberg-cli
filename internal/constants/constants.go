package constants

const (
	AppName             = "cb"
	ConfigDirName       = "cb"
	ConfigFileName      = "config.json"
	UpdateRepository    = "rishabyd/codeberg-cli"
	ReleaseBinaryName   = "cb"
	ReleaseInstallDir   = "/usr/local/bin"
	CodebergHost        = "codeberg.org"
	CodebergBaseURL     = "https://" + CodebergHost
	CodebergAPIBaseURL  = CodebergBaseURL + "/api/v1"
	OAuthAuthorizeURL   = CodebergBaseURL + "/login/oauth/authorize"
	OAuthTokenURL       = CodebergBaseURL + "/login/oauth/access_token"
	OAuthClientID       = "b8c299ec-4982-47db-afaa-e57972007216"
	OAuthCallbackHost   = "localhost"
	OAuthCallbackPort   = 3214
	OAuthRedirectURI    = "http://localhost:3214/callback"
	GitCredentialTarget = "credential.https://codeberg.org.helper"
	DefaultMigrateService = "github"
)

var StatusBadgeIDs = []int{1, 38, 29}
