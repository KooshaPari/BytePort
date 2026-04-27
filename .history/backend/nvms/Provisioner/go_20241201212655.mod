module github.com/provisioner

go 1.21

toolchain go1.23.2

require github.com/fermyon/spin/sdk/go/v2 v2.2.0

require (
	github.com/google/go-github/v67/github v67.0.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
)

require nvms v0.0.0

replace nvms => ../../nvms
