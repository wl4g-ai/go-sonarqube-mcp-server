//go:build darwin

package mcputils

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

//	getFromKeychain retrieves the SonarQube token from macOS Keychain.
//	Returns "" on any error, timeout, or panic — this is purely optional.
//
//	To store a token in Keychain:
//
//		security add-generic-password -s sonarqube-mcp -a sonarqube-mcp -w <your-token>
//
//	Customize the service name:
//
//		export SONARQUBE_KEYCHAIN_SERVICE=sonarqube-mcp
func getFromKeychain() (token string) {
	defer func() { recover() }()
	service := "sonarqube-mcp"
	if s := os.Getenv("SONARQUBE_KEYCHAIN_SERVICE"); s != "" {
		service = s
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "security", "find-generic-password", "-s", service, "-wa", "")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getFromWinCred() string { return "" }
