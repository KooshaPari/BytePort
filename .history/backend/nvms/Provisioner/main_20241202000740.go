// go.mod
module github.com/yourusername/byteport/backend/nvms

go 1.21  // or your Go version

require (
    github.com/go-git/go-billy/v5 v5.5.0
    // Remove Unix-specific dependencies
)

// Add replace directive for WASI compatibility
replace github.com/go-git/go-billy/v5 => github.com/go-git/go-billy/v5 v5.5.0