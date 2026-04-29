package oauth

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/constants"
)

func WaitForAuthCode(ctx context.Context, expectedState string) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if e := q.Get("error"); e != "" {
			errCh <- fmt.Errorf("oauth error: %s", e)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}
		if state := q.Get("state"); state != expectedState {
			errCh <- fmt.Errorf("csrf state mismatch")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>Authentication complete</title><style>body{margin:0;min-height:100vh;display:grid;place-items:center;background:linear-gradient(180deg,#f8fafc,#eef2ff);color:#0f172a;font-family:\"IBM Plex Sans\",\"Segoe UI\",Tahoma,sans-serif}.card{max-width:560px;margin:24px;padding:28px 24px;border:1px solid #dbeafe;border-radius:14px;background:#fff;box-shadow:0 12px 30px rgba(15,23,42,.08)}h1{margin:0 0 10px;font-size:clamp(1.3rem,2.8vw,1.8rem)}p{margin:0;font-size:1rem;line-height:1.55;color:#334155}</style></head><body><main class=\"card\"><h1>Authentication complete</h1><p>You can close this tab and return to your terminal.</p></main><script>setTimeout(function(){window.close()},250)</script></body></html>")
		codeCh <- html.EscapeString(code)
	})

	addr := fmt.Sprintf("%s:%d", constants.OAuthCallbackHost, constants.OAuthCallbackPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case code := <-codeCh:
		return code, nil
	}
}
