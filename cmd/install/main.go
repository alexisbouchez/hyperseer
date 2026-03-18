package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/alexisbouchez/hyperseer/internal/exit"
)

type installConfig struct {
	Domain     string
	CHPassword string
	Retention  string
	CaddyMode  string // "new", "existing", "skip"
}

func randomPassword() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func checkDocker() error {
	if err := exec.Command("docker", "info").Run(); err != nil {
		return fmt.Errorf("docker is not available — is Docker running?")
	}
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		return fmt.Errorf("docker compose plugin is not available")
	}
	return nil
}

func portFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func main() {
	var (
		flagDomain    = flag.String("domain", "", "Domain (e.g. otel.example.com)")
		flagPassword  = flag.String("password", "", "ClickHouse password (auto-generated if omitted)")
		flagRetention = flag.String("retention", "", "Data retention in days: 7, 30, 90, or 365")
		flagCaddy     = flag.String("caddy", "", "Proxy mode: new, existing, or skip")
		flagYes       = flag.Bool("yes", false, "Skip confirmation prompt")
	)
	flag.Parse()

	// Track which flags were explicitly set
	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })

	fmt.Println("\033[1mHyperseer Installation\033[0m")
	fmt.Println()

	fmt.Print("Checking Docker… ")
	if err := checkDocker(); err != nil {
		fmt.Println("\033[31m✗\033[0m")
		exit.WithError(err)
	}
	fmt.Println("\033[32m✓\033[0m")
	fmt.Println()

	cfg := installConfig{
		CHPassword: randomPassword(),
		Retention:  "30",
		CaddyMode:  "new",
	}
	if set["domain"] {
		cfg.Domain = *flagDomain
	}
	if set["password"] {
		cfg.CHPassword = *flagPassword
	}
	if set["retention"] {
		cfg.Retention = *flagRetention
	}
	if set["caddy"] {
		cfg.CaddyMode = *flagCaddy
	}

	// Group 1: domain, password, retention — only fields not provided via flags
	var group1 []huh.Field
	if !set["domain"] {
		group1 = append(group1, huh.NewInput().
			Title("Domain").
			Description("Where Hyperseer will be accessible (e.g. otel.example.com)").
			Value(&cfg.Domain))
	}
	if !set["password"] {
		group1 = append(group1, huh.NewInput().
			Title("ClickHouse password").
			Value(&cfg.CHPassword))
	}
	if !set["retention"] {
		group1 = append(group1, huh.NewSelect[string]().
			Title("Data retention").
			Options(
				huh.NewOption("7 days", "7"),
				huh.NewOption("30 days", "30"),
				huh.NewOption("90 days", "90"),
				huh.NewOption("1 year", "365"),
			).
			Value(&cfg.Retention))
	}
	if len(group1) > 0 {
		if err := huh.NewForm(huh.NewGroup(group1...)).Run(); err != nil {
			exit.WithError(err)
		}
	}

	// Group 2: caddy mode
	if !set["caddy"] {
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Reverse proxy / HTTPS").
					Options(
						huh.NewOption("Create a new Caddy instance (Docker)", "new"),
						huh.NewOption("I have an existing Caddy — show me the config snippet", "existing"),
						huh.NewOption("Skip proxy configuration", "skip"),
					).
					Value(&cfg.CaddyMode),
			),
		).Run(); err != nil {
			exit.WithError(err)
		}
	}

	if cfg.CaddyMode == "new" {
		fmt.Print("Checking ports 80 and 443… ")
		var busy []string
		if !portFree(80) {
			busy = append(busy, "80")
		}
		if !portFree(443) {
			busy = append(busy, "443")
		}
		if len(busy) > 0 {
			fmt.Println("\033[31m✗\033[0m")
			exit.WithError(fmt.Errorf("port(s) %s already in use — free them or choose a different proxy option", strings.Join(busy, ", ")))
		}
		fmt.Println("\033[32m✓\033[0m")
		fmt.Println()
	}

	// Confirmation
	confirm := *flagYes
	if !confirm {
		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Generate configuration files and start services?").
					Affirmative("Yes, let's go").
					Negative("Cancel").
					Value(&confirm),
			),
		).Run(); err != nil {
			exit.WithError(err)
		}
	}
	if !confirm {
		fmt.Println("Aborted.")
		return
	}

	if err := writeEnv(cfg); err != nil {
		exit.WithError(err)
	}
	fmt.Println("\033[32m•\033[0m wrote .env")

	if err := writeCompose(cfg); err != nil {
		exit.WithError(err)
	}
	fmt.Println("\033[32m•\033[0m wrote docker-compose.yml")

	switch cfg.CaddyMode {
	case "new":
		if err := writeCaddyfile(cfg); err != nil {
			exit.WithError(err)
		}
		fmt.Println("\033[32m•\033[0m wrote Caddyfile")
	case "existing":
		fmt.Printf("\n\033[1mAdd this block to your Caddyfile:\033[0m\n\n%s\n", caddySnippet(cfg))
	}

	fmt.Println()
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		exit.WithError(fmt.Errorf("docker compose up: %w", err))
	}

	fmt.Printf("\n\033[32m•\033[0m Hyperseer is live")
	if cfg.Domain != "" {
		fmt.Printf(" at https://%s", cfg.Domain)
	}
	fmt.Println()
}

func writeEnv(cfg installConfig) error {
	content := fmt.Sprintf("CH_PASSWORD=%s\nDOMAIN=%s\nRETENTION_DAYS=%s\n",
		cfg.CHPassword, cfg.Domain, cfg.Retention)
	return os.WriteFile(".env", []byte(content), 0600)
}

func writeCompose(cfg installConfig) error {
	var sb strings.Builder

	sb.WriteString(`services:
  clickhouse:
    image: clickhouse/clickhouse-server:25.3
    environment:
      CLICKHOUSE_DB: hyperseer
      CLICKHOUSE_USER: hyperseer
      CLICKHOUSE_PASSWORD: ${CH_PASSWORD}
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    restart: unless-stopped

  serve:
    image: ghcr.io/alexisbouchez/hyperseer-serve:latest
    environment:
      HYPERSEER_CH_HOST: clickhouse
      HYPERSEER_CH_PASSWORD: ${CH_PASSWORD}
    depends_on:
      - clickhouse
    restart: unless-stopped
`)

	if cfg.CaddyMode == "new" {
		sb.WriteString(`
  caddy:
    image: caddy:2
    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - serve
    restart: unless-stopped
`)
	}

	sb.WriteString(`
volumes:
  clickhouse_data:
`)
	if cfg.CaddyMode == "new" {
		sb.WriteString("  caddy_data:\n  caddy_config:\n")
	}

	return os.WriteFile("docker-compose.yml", []byte(sb.String()), 0644)
}

func caddySnippet(cfg installConfig) string {
	return fmt.Sprintf("%s {\n\treverse_proxy serve:4318\n}\n", cfg.Domain)
}

func writeCaddyfile(cfg installConfig) error {
	return os.WriteFile("Caddyfile", []byte(caddySnippet(cfg)), 0644)
}
