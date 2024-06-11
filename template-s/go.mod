module github.com/nesiler/cestx/template-s

go 1.22.3

// for local development
replace github.com/nesiler/cestx/common => ../common

replace github.com/nesiler/cestx/minio => ../minio

replace github.com/nesiler/cestx/redis => ../redis

replace github.com/nesiler/cestx/rabbitmq => ../rabbitmq

replace github.com/nesiler/cestx/postgresql => ../postgresql

replace github.com/nesiler/cestx/postgresql/models => ../postgresql/models

require (
	github.com/nesiler/cestx/common v0.0.0-20240610110910-eae36ab76517
	github.com/nesiler/cestx/minio v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/streadway/amqp v1.1.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/rs/xid v1.5.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/postgres v1.5.7 // indirect
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/minio-go/v7 v7.0.71
	github.com/nesiler/cestx/postgresql v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/postgresql/models v0.0.0-20240610191142-1ba8680a3682
	golang.org/x/sys v0.18.0 // indirect
	gorm.io/gorm v1.25.10
)
