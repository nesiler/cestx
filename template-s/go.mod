module github.com/nesiler/cestx/template-s

go 1.22.3

// for local development
replace github.com/nesiler/cestx/common => ../common

replace github.com/nesiler/cestx/minio => ../minio

replace github.com/nesiler/cestx/redis => ../redis

replace github.com/nesiler/cestx/rabbitmq => ../rabbitmq

require (
	github.com/nesiler/cestx/common v0.0.0-20240610110910-eae36ab76517
	github.com/nesiler/cestx/minio v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/redis v0.0.0-00010101000000-000000000000
	github.com/streadway/amqp v1.1.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/rs/xid v1.5.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/minio-go/v7 v7.0.71 // indirect
	github.com/redis/go-redis/v9 v9.5.2 // indirect
	golang.org/x/sys v0.18.0 // indirect
)
