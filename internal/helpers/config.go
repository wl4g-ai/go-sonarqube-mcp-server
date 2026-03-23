package mcputils

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
)

type contextKey string

const HTTPHeadersContextKey contextKey = "mcp-http-headers"

type SQMode int

const (
	SQModeServer SQMode = iota
	SQModeCloud
)

var (
	configOnce   sync.Once
	sonarQubeURL string
	sqToken      string
	sqOrg        string
	sqProjectKey string
	sqToolSets   string
	sqReadOnly   bool
	sqMode       SQMode
)

func initConfig() {
	sonarQubeURL = os.Getenv("SONARQUBE_URL")
	sqToken = resolveToken()
	sqOrg = os.Getenv("SONARQUBE_ORG")
	sqProjectKey = os.Getenv("SONARQUBE_PROJECT_KEY")
	sqToolSets = os.Getenv("SONARQUBE_TOOLSETS")
	sqReadOnly = strings.ToLower(os.Getenv("SONARQUBE_READ_ONLY")) == "true"

	if sqOrg != "" {
		sqMode = SQModeCloud
		if sonarQubeURL == "" {
			sonarQubeURL = "https://sonarcloud.io"
		}
	} else {
		sqMode = SQModeServer
	}
}

func resolveToken() string {
	// 1. SONARQUBE_TOKEN env var
	if t := os.Getenv("SONARQUBE_TOKEN"); t != "" {
		return t
	}
	// 2. SONARQUBE_TOKEN_FILE (for Docker/Kubernetes secrets)
	if f := os.Getenv("SONARQUBE_TOKEN_FILE"); f != "" {
		if data, err := os.ReadFile(f); err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	// 3. Fall back to keychain / wincred (reuses existing helpers)
	if t := getFromKeychain(); t != "" {
		return t
	}
	if t := getFromWinCred(); t != "" {
		return t
	}
	return ""
}

// GetSonarQubeURL returns the SonarQube instance URL.
func GetSonarQubeURL() string {
	configOnce.Do(initConfig)
	return sonarQubeURL
}

// GetSonarQubeToken returns the SonarQube auth token.
func GetSonarQubeToken() string {
	configOnce.Do(initConfig)
	return sqToken
}

// GetSonarQubeOrg returns the SonarQube Cloud organization key, empty for Server.
func GetSonarQubeOrg() string {
	configOnce.Do(initConfig)
	return sqOrg
}

// GetDefaultProjectKey returns the fallback project key, if configured.
func GetDefaultProjectKey() string {
	configOnce.Do(initConfig)
	return sqProjectKey
}

// GetToolSets returns the configured toolset filter, or "" for all.
func GetToolSets() string {
	configOnce.Do(initConfig)
	return sqToolSets
}

// IsReadOnly returns true when mutation tools should be disabled.
func IsReadOnly() bool {
	configOnce.Do(initConfig)
	return sqReadOnly
}

// IsCloud returns true when connected to SonarQube Cloud (SONARQUBE_ORG is set).
func IsCloud() bool {
	configOnce.Do(initConfig)
	return sqMode == SQModeCloud
}

// ToolSetEnabled checks if a given category is included in SONARQUBE_TOOLSETS.
// An empty toolset string means all toolsets are enabled.
// WithHTTPHeaders stores HTTP headers in the context for forwarding to upstream.
func WithHTTPHeaders(ctx context.Context, headers http.Header) context.Context {
	return context.WithValue(ctx, HTTPHeadersContextKey, headers)
}

// GetHTTPHeaders retrieves HTTP headers from the context.
func GetHTTPHeaders(ctx context.Context) http.Header {
	if h, ok := ctx.Value(HTTPHeadersContextKey).(http.Header); ok {
		return h
	}
	return nil
}

func ToolSetEnabled(category string) bool {
	configOnce.Do(initConfig)
	if sqToolSets == "" {
		return true
	}
	for _, t := range strings.Split(sqToolSets, ",") {
		if strings.EqualFold(strings.TrimSpace(t), category) {
			return true
		}
	}
	return false
}
