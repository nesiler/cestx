module github.com/nesiler/cestx/postgresql

go 1.22.3

require (
	github.com/google/uuid v1.6.0
	github.com/nesiler/cestx/common v0.0.0-20240605091303-10941c2ebc65
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
)

require github.com/joho/godotenv v1.5.1 // indirect

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nesiler/cestx/postgresql/models v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

// for local development
replace github.com/nesiler/cestx/common => ../common

replace github.com/nesiler/cestx/postgresql/models => ./models
