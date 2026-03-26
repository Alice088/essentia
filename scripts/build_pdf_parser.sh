#!/bin/bash

INPUT=${1}
OUTPUT=${2:-pdf-parser}

if [ -z "$INPUT" ]; then
  echo "usage: $0 <input_path> [output_binary]"
  exit 1
fi

#exmaple ./scripts/build_pdf_parser.sh ./cmd/pdf_parser ./build
# ./scripts/build_pdf_parser.sh - script
# ./cmd/pdf_parser - path to go bi
# ./build - path to build
# output: ./build/pdf_parser
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "$OUTPUT" "$INPUT"

echo "success built"