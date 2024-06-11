module github.com/nesiler/cestx/machine-s

go 1.22.3

// for local development
replace github.com/nesiler/cestx/common => ../common

replace github.com/nesiler/cestx/rabbitmq => ../rabbitmq

replace github.com/nesiler/cestx/minio => ../minio

replace github.com/nesiler/cestx/postgresql => ../postgresql

replace github.com/nesiler/cestx/redis => ../redis

require (
	github.com/docker/docker v26.1.4+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/minio/minio-go/v7 v7.0.71
	github.com/nesiler/cestx/common v0.0.0-20240605091303-10941c2ebc65
	github.com/nesiler/cestx/minio v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/postgresql v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/postgresql/models v0.0.0-20240610191142-1ba8680a3682
	github.com/nesiler/cestx/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/nesiler/cestx/redis v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.5.3
	github.com/streadway/amqp v1.1.0
	gorm.io/gorm v1.25.10
)

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/rs/xid v1.5.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/postgres v1.5.7 // indirect
	gotest.tools/v3 v3.5.1 // indirect
)
