package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/alexisbouchez/hyperseer/internal/exit"
)

type authConfigResponse struct {
	Provider string `json:"provider"`
	URL      string `json:"url"`
	Realm    string `json:"realm"`
	ClientID string `json:"client_id"`
}

func fetchAuthConfig(server string) (authConfigResponse, error) {
	resp, err := http.Get(server + "/auth/config")
	if err != nil {
		return authConfigResponse{}, fmt.Errorf("could not reach server: %w", err)
	}
	defer resp.Body.Close()
	var cfg authConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return authConfigResponse{}, fmt.Errorf("invalid auth config response: %w", err)
	}
	return cfg, nil
}

var loginCommand = &cli.Command{
	Name:  "login",
	Usage: "Authenticate with your Hyperseer instance",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		server := serverURL(cmd)
		fmt.Printf("\033[2mfetching auth config from %s…\033[0m\n", server)

		cfg, err := fetchAuthConfig(server)
		if err != nil {
			return err
		}

		switch cfg.Provider {
		case "keycloak":
			return keycloakBrowserLogin(cfg.URL, cfg.Realm, cfg.ClientID)
		case "supabase":
			return supabaseBrowserLogin(cfg.URL)
		default:
			return fmt.Errorf("server returned unknown provider %q", cfg.Provider)
		}
	},
}

// pkce returns a random code_verifier and its S256 code_challenge.
func pkce() (verifier, challenge string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func listenLocalhost() (net.Listener, int) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		exit.WithError(err)
	}
	return ln, ln.Addr().(*net.TCPAddr).Port
}

func openBrowser(rawURL string) {
	fmt.Printf("\033[2m  opening %s\033[0m\n", rawURL)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "linux":
		cmd = exec.Command("xdg-open", rawURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", rawURL)
	default:
		fmt.Printf("  visit: %s\n", rawURL)
		return
	}
	_ = cmd.Start()
}

func keycloakBrowserLogin(baseURL, realm, clientID string) error {
	verifier, challenge := pkce()
	ln, port := listenLocalhost()
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?%s", baseURL, realm, params.Encode())

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if e := r.URL.Query().Get("error"); e != "" {
			desc := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("%s: %s", e, desc)
			fmt.Fprintln(w, `<html><body><p>Authentication failed. You can close this tab.</p></body></html>`)
			return
		}
		codeCh <- r.URL.Query().Get("code")
		fmt.Fprintln(w, `<html><body><p>Logged in! You can close this tab.</p></body></html>`)
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	openBrowser(authURL)
	fmt.Println("waiting for browser…")

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("login timed out")
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", baseURL, realm)
	resp, err := http.PostForm(tokenURL, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	accessToken, expiresIn, err := parseTokenResponse(resp)
	if err != nil {
		return err
	}
	if err := saveToken(storedToken{
		AccessToken: accessToken,
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}); err != nil {
		return err
	}

	fmt.Println("\033[32m•\033[0m logged in")
	return nil
}

func supabaseBrowserLogin(baseURL string) error {
	ln, port := listenLocalhost()
	doneCh := make(chan error, 1)

	loginHTML := `<!DOCTYPE html>
<html><head><title>Hyperseer Login</title><style>
body{font-family:system-ui,sans-serif;max-width:360px;margin:80px auto;padding:0 20px}
h2{margin-bottom:24px}
label{font-size:13px;font-weight:600}
input{display:block;width:100%;margin:6px 0 16px;padding:9px 10px;font-size:14px;border:1px solid #d1d5db;border-radius:6px;box-sizing:border-box}
button{background:#0f172a;color:#fff;border:none;padding:10px 20px;font-size:14px;border-radius:6px;cursor:pointer;width:100%}
#err{color:#dc2626;font-size:13px;margin-top:12px;min-height:18px}
</style></head><body>
<h2>Hyperseer</h2>
<form id="f">
  <label>Email</label><input type="email" id="email" required autofocus>
  <label>Password</label><input type="password" id="password" required>
  <button type="submit">Sign in</button>
  <p id="err"></p>
</form>
<script>
document.getElementById('f').addEventListener('submit',async e=>{
  e.preventDefault();
  const errEl=document.getElementById('err');
  errEl.textContent='';
  const r=await fetch('/login',{method:'POST',headers:{'Content-Type':'application/json'},
    body:JSON.stringify({email:document.getElementById('email').value,password:document.getElementById('password').value})});
  if(!r.ok){errEl.textContent=await r.text();return;}
  document.body.innerHTML='<p style="margin-top:80px;text-align:center">Logged in! You can close this tab.</p>';
});
</script>
</body></html>
`

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, loginHTML)
	})
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		accessToken, expiresIn, err := supabaseLogin(baseURL, creds.Email, creds.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := saveToken(storedToken{
			AccessToken: accessToken,
			ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		doneCh <- nil
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	openBrowser(fmt.Sprintf("http://localhost:%d", port))
	fmt.Println("waiting for browser…")

	select {
	case err := <-doneCh:
		if err != nil {
			return err
		}
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("login timed out")
	}

	fmt.Println("\033[32m•\033[0m logged in")
	return nil
}

func supabaseLogin(baseURL, email, password string) (string, int, error) {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := http.Post(
		baseURL+"/token?grant_type=password",
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	return parseTokenResponse(resp)
}

func parseTokenResponse(resp *http.Response) (string, int, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("auth failed (%s): %s", resp.Status, body)
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, err
	}
	return result.AccessToken, result.ExpiresIn, nil
}
