module github.com/nesiler/cestx/redis

go 1.22.3

require github.com/redis/go-redis/v9 v9.5.2

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0
	github.com/nesiler/cestx/common v0.0.0-20240605091303-10941c2ebc65
)

// for local development
replace github.com/nesiler/cestx/common => ../common
