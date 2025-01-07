module parser

go 1.22.2

toolchain go1.23.2

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
)

require (
	github.com/fermyon/spin/sdk/go/v2 v2.2.0
	gopkg.in/yaml.v3 v3.0.1
	nvms v0.0.0
)

replace nvms => ../../nvms
