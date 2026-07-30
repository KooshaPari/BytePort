//go:build tools

package tools

// Keep govulncheck in the locked tools module so CI installs the audited version.
import _ "golang.org/x/vuln/cmd/govulncheck"
