#!/bin/bash

services=(
  "registry"
  "logger-s"
  "taskmaster-s"
  "dynoxy-s"
  "template-s"
  "machine-s"
  "api-gw"
  "redis"
  "postgresql"
  "rabbitmq"
)

build_commands=(
  "go build -o registry && ./registry"
  "go build -o logger-s && ./logger-s"
  "go build -o taskmaster-s && ./taskmaster-s"
  "go build -o dynoxy-s && ./dynoxy-s"
  "go build -o template-s && ./template-s"
  "go build -o machine-s && ./machine-s"
  "dotnet run"
  "docker-compose up -d"
  "docker-compose up -d"
  "docker-compose up -d"
)

for i in "${!services[@]}"; do
  service=${services[$i]}
  build_command=${build_commands[$i]}

  cat <<EOF > ${service}.service
[Unit]
Description=${service} Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/cestx/${service}
ExecStart=${build_command}
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

done

echo "Service files created successfully."