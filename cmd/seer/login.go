package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"

	"github.com/alexisbouchez/hyperseer/internal/exit"
)

var loginCommand = &cli.Command{
	Name:  "login",
	Usage: "Authenticate with your Hyperseer instance",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "provider", Aliases: []string{"p"}, Usage: "Auth provider: keycloak or supabase"},
		&cli.StringFlag{Name: "url", Usage: "Provider base URL (e.g. http://localhost:8080)"},
		&cli.StringFlag{Name: "realm", Value: "hyperseer", Usage: "Keycloak realm"},
		&cli.StringFlag{Name: "client-id", Value: "hyperseer-cli", Usage: "Keycloak client ID"},
		&cli.StringFlag{Name: "email", Usage: "Email / username"},
		&cli.StringFlag{Name: "password", Usage: "Password"},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		provider := cmd.String("provider")
		providerURL := cmd.String("url")
		email := cmd.String("email")
		password := cmd.String("password")

		// Prompt for anything missing
		var formFields []huh.Field
		if provider == "" {
			formFields = append(formFields, huh.NewSelect[string]().
				Title("Provider").
				Options(
					huh.NewOption("Keycloak", "keycloak"),
					huh.NewOption("Supabase", "supabase"),
				).
				Value(&provider))
		}
		if providerURL == "" {
			formFields = append(formFields, huh.NewInput().
				Title("Provider URL").
				Value(&providerURL))
		}
		if email == "" {
			formFields = append(formFields, huh.NewInput().
				Title("Email").
				Value(&email))
		}
		if password == "" {
			formFields = append(formFields, huh.NewInput().
				Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password))
		}
		if len(formFields) > 0 {
			if err := huh.NewForm(huh.NewGroup(formFields...)).Run(); err != nil {
				exit.WithError(err)
			}
		}

		var (
			accessToken string
			expiresIn   int
			err         error
		)

		switch provider {
		case "keycloak":
			accessToken, expiresIn, err = keycloakLogin(
				providerURL, cmd.String("realm"), cmd.String("client-id"), email, password,
			)
		case "supabase":
			accessToken, expiresIn, err = supabaseLogin(providerURL, email, password)
		default:
			return fmt.Errorf("unknown provider %q — use keycloak or supabase", provider)
		}
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
	},
}

func keycloakLogin(baseURL, realm, clientID, username, password string) (string, int, error) {
	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", baseURL, realm)
	resp, err := http.PostForm(endpoint, url.Values{
		"grant_type": {"password"},
		"client_id":  {clientID},
		"username":   {username},
		"password":   {password},
	})
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	return parseTokenResponse(resp)
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
