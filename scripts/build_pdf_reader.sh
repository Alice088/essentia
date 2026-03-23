#!/bin/bash

INPUT=${1}
OUTPUT=${2:-pdf-reader}

if [ -z "$INPUT" ]; then
  echo "usage: $0 <input_path> [output_binary]"
  exit 1
fi

#exmaple ./scripts/build_pdf_reader.sh ./cmd/pdf_reader ./build
# ./scripts/build_pdf_reader.sh - script
# ./cmd/pdf_reader - path to go bi
# ./build - path to build
# output: ./build/pdf_reader
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "$OUTPUT" "$INPUT"

echo "success built"