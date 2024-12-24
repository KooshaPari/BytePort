module github.com/demonstrator

go 1.22.2

toolchain go1.23.2

require github.com/fermyon/spin/sdk/go/v2 v2.2.0

require (
	aidanwoods.dev/go-paseto v1.5.2 // indirect
	aidanwoods.dev/go-result v0.1.0 // indirect
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/fermyon/spin-go-sdk v0.0.0-20240918180601-c2d4ef2e0904 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
)

require (
	github.com/openai/openai-go v0.1.0-alpha.41
	nvms v0.0.0
)

replace nvms => ../../nvms
