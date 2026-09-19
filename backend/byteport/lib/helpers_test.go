package lib

import "os"

// osGetenv is os.Getenv, factored into a helper so test files don't need to
// import "os" just to read an environment variable.
func osGetenv(key string) string {
	return os.Getenv(key)
}
