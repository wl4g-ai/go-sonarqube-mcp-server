//go:build windows

package mcputils

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

//	getFromWinCred retrieves the SonarQube token from Windows Credential Manager.
//	Returns "" on any error, timeout, or panic — this is purely optional.
//
//	To store a token:
//
//		cmdkey /add:sonarqube-mcp /user:sonarqube-mcp /pass:your-token
//
//	Customize the target name:
//
//		set SONARQUBE_WINCRED_TARGET=sonarqube-mcp
func getFromWinCred() (token string) {
	defer func() { recover() }()
	target := "sonarqube-mcp"
	if t := os.Getenv("SONARQUBE_WINCRED_TARGET"); t != "" {
		target = t
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "cmdkey", "/get:"+target)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Password:") {
			return strings.TrimPrefix(line, "Password:")
		}
	}
	return ""
}

func getFromKeychain() string { return "" }
