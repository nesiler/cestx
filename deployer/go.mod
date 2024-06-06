module github.com/nesiler/cestx/deployer

go 1.22.3

require (
	github.com/google/go-github/v35 v35.3.0
	github.com/nesiler/cestx/common v0.0.0-20240529184331-b25e403c4f6e
	golang.org/x/oauth2 v0.20.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robfig/cron/v3 v3.0.1
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 // indirect
	golang.org/x/sys v0.18.0 // indirect
)

// for local development
replace github.com/nesiler/cestx/common => ../common
