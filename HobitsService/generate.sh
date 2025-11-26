#!/bin/bash

# Генерирует Go код из proto файлов

PROTO_DIR="proto"
GEN_DIR="gen/go"

# Создаем директорию для сгенерированного кода
mkdir -p "$GEN_DIR"

# Генерируем код для каждого proto файла
protoc --go_out="$GEN_DIR" --go-grpc_out="$GEN_DIR" \
  -I"$PROTO_DIR" \
  "$PROTO_DIR"/common.proto \
  "$PROTO_DIR"/user_service.proto \
  "$PROTO_DIR"/habit_service.proto \
  "$PROTO_DIR"/log_service.proto \
  "$PROTO_DIR"/reminder_service.proto

echo "Proto files generated successfully!"
echo "Generated files are in: $GEN_DIR"
