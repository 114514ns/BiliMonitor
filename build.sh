#!/bin/bash


APP_NAME="BiliMonitor"

OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"


declare -A TARGETS=(
  ["windows_amd64"]="GOOS=windows GOARCH=amd64"
  ["linux_amd64"]="GOOS=linux GOARCH=amd64"
  ["linux_mips64"]="GOOS=linux GOARCH=mips64 CGO_ENABLED=1 CC=mips64-linux-gnu-gcc"
  ["linux_arm64"]="GOOS=linux GOARCH=arm64"
  ["darwin_amd64"]="GOOS=darwin GOARCH=amd64"
  ["darwin_arm64"]="GOOS=darwin GOARCH=arm64"
)

for target in "${!TARGETS[@]}"; do
  eval "${TARGETS[$target]}" go build -o "$OUTPUT_DIR/$APP_NAME-$target" .
  if [ $? -eq 0 ]; then
    echo "Build successful: $OUTPUT_DIR/$APP_NAME-$target"
    if [[ "$target" != darwin_* ]]; then
        upx --best --lzma "$OUTPUT_DIR/$APP_NAME-$target"
    fi
  else
    echo "Build failed: $target"
  fi
done

