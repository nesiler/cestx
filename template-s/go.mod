module github.com/nesiler/cestx/template-s

go 1.22.3

// for local development
replace github.com/nesiler/cestx/common => ../common

require (
	github.com/joho/godotenv v1.5.1
	github.com/nesiler/cestx/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
