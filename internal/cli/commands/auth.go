package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/g0ulartleo/mirante/internal/cli"
	"github.com/g0ulartleo/mirante/internal/config"
)

type AuthCommand struct{}

func (a *AuthCommand) Name() string {
	return "auth"
}

func (a *AuthCommand) Description() string {
	return "Authenticate with the Mirante server using OAuth"
}

func (a *AuthCommand) Usage() string {
	return "auth <api_host>"
}

func (a *AuthCommand) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ./cli auth <api_host>")
	}

	apiHost := args[0]

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	cliConfig.APIHost = apiHost

	if !checkOAuthSupport(apiHost) {
		fmt.Println("Server does not support OAuth. Please use the 'auth-key' command with an API key.")
		return nil
	}

	fmt.Println("Opening browser for authentication...")

	loginURL := fmt.Sprintf("%s/auth/login", apiHost)
	if err := openBrowser(loginURL); err != nil {
		fmt.Printf("Failed to open browser. Please manually visit: %s\n", loginURL)
	}

	fmt.Println("After completing authentication in the browser, you'll receive a token.")
	fmt.Print("Please paste the token here: ")

	var token string
	if _, err := fmt.Scanln(&token); err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	cliConfig.AuthToken = token
	cliConfig.AuthType = "oauth"

	if !validateToken(apiHost, token) {
		return fmt.Errorf("invalid token provided")
	}

	if err := config.SaveCLIConfig(cliConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Authentication successful! OAuth token saved.")
	return nil
}

func checkOAuthSupport(apiHost string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/auth/status", apiHost))
	if err != nil {
		fmt.Println("Error checking OAuth support:", err)
		return false
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.StatusCode)

	return resp.StatusCode != 404
}

func validateToken(apiHost, token string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/auth/status", apiHost), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}
	fmt.Println("token", token)

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	authenticated, ok := result["authenticated"].(bool)
	return ok && authenticated
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func init() {
	a := &AuthCommand{}
	cli.RegisterCommand(a.Name(), a)
}
