#!/bin/bash


APP_NAME="BiliMonitor"

OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"


declare -A TARGETS=(
  ["windows_amd64"]="GOOS=windows GOARCH=amd64"
  ["windows_arm64"]="GOOS=windows GOARCH=arm64"
  ["linux_amd64"]="GOOS=linux GOARCH=amd64"
  ["linux_arm64"]="GOOS=linux GOARCH=arm64"
  ["darwin_amd64"]="GOOS=darwin GOARCH=amd64"
  ["darwin_arm64"]="GOOS=darwin GOARCH=arm64"
  ["android_arm64"]="GOOS=android GOARCH=arm64"
)

for target in "${!TARGETS[@]}"; do
  eval "${TARGETS[$target]}" go build -ldflags="-s" -o "$OUTPUT_DIR/$APP_NAME-$target"
  # eval "${TARGETS[$target]}"
  # echo "Building for $target: GOOS=${GOOS} GOARCH=${GOARCH}"
  # go build -o "$OUTPUT_DIR/$APP_NAME-$target" -ldflags="-s -w"
  if [ $? -eq 0 ]; then
    echo "Build successful: $OUTPUT_DIR/$APP_NAME-$target"
    if [[ "$target" != darwin_* ]]; then
        upx --best --lzma "$OUTPUT_DIR/$APP_NAME-$target"
    fi
  else
    echo "Build failed: $target"
  fi
done

