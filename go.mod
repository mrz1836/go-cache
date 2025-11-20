module github.com/mrz1836/go-cache

go 1.24.0

require (
	github.com/gomodule/redigo v1.9.3
	github.com/newrelic/go-agent/v3 v3.42.0
	github.com/rafaeljusto/redigomock v2.4.0+incompatible
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Security: Force golang.org/x/crypto to v0.45.0 to fix CVE-2025-47914 and CVE-2025-58181
replace golang.org/x/crypto => golang.org/x/crypto v0.45.0
