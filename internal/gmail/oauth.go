package gmail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Required least-privilege Gmail OAuth scopes.
var RequiredScopes = []string{
	gmail.GmailSendScope,
	gmail.GmailReadonlyScope,
}

// OAuthConfig holds the OAuth client credentials and paths.
type OAuthConfig struct {
	CredentialsFile string
	TokenFile       string
	LocalPort       int
}

// CredentialsFileStructure parses GCP OAuth client secret JSON.
type CredentialsFileStructure struct {
	Installed *struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
	} `json:"installed"`
	Web *struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
	} `json:"web"`
}

// LoadOAuth2Config reads client credentials from disk and sets up the oauth2.Config.
func LoadOAuth2Config(credentialsPath string, localPort int) (*oauth2.Config, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file %s: %w", credentialsPath, err)
	}

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", localPort)

	// Try standard google.ConfigFromJSON
	cfg, err := google.ConfigFromJSON(data, RequiredScopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Google OAuth credentials: %w", err)
	}

	cfg.RedirectURL = redirectURI
	return cfg, nil
}

// ReadSavedToken reads and deserializes the stored OAuth2 token from disk.
func ReadSavedToken(tokenPath string) (*oauth2.Token, error) {
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tok oauth2.Token
	if err := json.NewDecoder(f).Decode(&tok); err != nil {
		return nil, fmt.Errorf("failed to decode token JSON: %w", err)
	}
	return &tok, nil
}

// SaveToken writes the OAuth2 token securely with 0600 file permissions.
func SaveToken(tokenPath string, tok *oauth2.Token) error {
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	// Write with restrictive 0600 permissions
	return os.WriteFile(tokenPath, data, 0600)
}

// GetTokenSource returns an automatically refreshing TokenSource using the saved token.
func GetTokenSource(ctx context.Context, oauthCfg *oauth2.Config, tokenPath string) (oauth2.TokenSource, error) {
	tok, err := ReadSavedToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("no saved token found: %w", err)
	}

	tokenSource := oauthCfg.TokenSource(ctx, tok)
	// Refresh check
	newTok, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// If token was refreshed, persist new token to disk
	if newTok.AccessToken != tok.AccessToken {
		_ = SaveToken(tokenPath, newTok)
	}

	return tokenSource, nil
}

// RunHeadlessOAuthHelper starts a local loopback server to receive the authorization code,
// while also providing instructions for SSH port forwarding or manual code entry.
func RunHeadlessOAuthHelper(ctx context.Context, oauthCfg *oauth2.Config, tokenPath string, localPort int, in io.Reader, out io.Writer) error {
	state := fmt.Sprintf("sparkwa-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local loopback listener
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return fmt.Errorf("failed to start local OAuth listener on port %d: %w", localPort, err)
	}
	defer listener.Close()

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/oauth/callback" {
				http.NotFound(w, r)
				return
			}

			if r.URL.Query().Get("state") != state {
				http.Error(w, "Invalid OAuth state token", http.StatusBadRequest)
				errChan <- errors.New("state mismatch in OAuth callback")
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				errMsg := r.URL.Query().Get("error")
				http.Error(w, "Authorization failed: "+errMsg, http.StatusBadRequest)
				errChan <- fmt.Errorf("authorization denied by Google: %s", errMsg)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif; text-align:center; padding:50px;">
  <h2 style="color:#2e7d32;">Authentication Successful!</h2>
  <p>SparkWA Lite has captured your authorization token.</p>
  <p>You can close this browser tab and return to your SSH terminal.</p>
</body>
</html>
`))
			codeChan <- code
		}),
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	fmt.Fprintln(out, "================================================================================")
	fmt.Fprintln(out, "               GOOGLE GMAIL OAUTH 2.0 HEADLESS SETUP")
	fmt.Fprintln(out, "================================================================================")
	fmt.Fprintln(out, "1. If you are on an SSH session from your laptop/desktop, forward port:")
	fmt.Fprintf(out, "   ssh -L %d:127.0.0.1:%d user@your-vps-ip\n\n", localPort, localPort)
	fmt.Fprintln(out, "2. Open this authorization URL in your browser:")
	fmt.Fprintln(out, authURL)
	fmt.Fprintln(out, "================================================================================")
	fmt.Fprintln(out, "Waiting for browser redirect on http://127.0.0.1:"+fmt.Sprint(localPort)+"/oauth/callback ...")
	fmt.Fprintln(out, "(Or paste the authorization code / redirect URL below if not forwarding ports)")
	fmt.Fprint(out, "Auth Code or Redirect URL [Press Enter to wait for browser]: ")

	// Allow manual pasting of auth code or full redirect URL as fallback
	manualCodeChan := make(chan string, 1)
	go func() {
		var input string
		_, _ = fmt.Fscanln(in, &input)
		input = strings.TrimSpace(input)
		if input != "" {
			// If user pasted full redirect URL: http://127.0.0.1:8085/oauth/callback?state=...&code=...
			if strings.Contains(input, "code=") {
				parts := strings.Split(input, "code=")
				if len(parts) > 1 {
					codeVal := parts[1]
					if amp := strings.Index(codeVal, "&"); amp != -1 {
						codeVal = codeVal[:amp]
					}
					manualCodeChan <- codeVal
					return
				}
			}
			manualCodeChan <- input
		}
	}()

	var authCode string
	select {
	case code := <-codeChan:
		authCode = code
		fmt.Fprintln(out, "\n✓ Received authorization code via loopback callback!")
	case code := <-manualCodeChan:
		authCode = code
		fmt.Fprintln(out, "\n✓ Received authorization code from terminal input!")
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Minute):
		return errors.New("timed out waiting for OAuth authorization (5 minute limit)")
	case <-ctx.Done():
		return ctx.Err()
	}

	_ = srv.Shutdown(context.Background())

	// Exchange authorization code for token
	token, err := oauthCfg.Exchange(ctx, authCode)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code for token: %w", err)
	}

	if err := SaveToken(tokenPath, token); err != nil {
		return fmt.Errorf("failed to save token to %s: %w", tokenPath, err)
	}

	fmt.Fprintf(out, "✓ Successfully saved OAuth token with refresh capabilities to %s (permissions 0600)\n", tokenPath)
	return nil
}
