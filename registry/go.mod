module github.com/nesiler/cestx/registry

go 1.22.3

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

// for local development
replace github.com/nesiler/cestx/registry => ../registry
