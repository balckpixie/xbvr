#!/bin/bash

# "go run" またはビルド済バイナリを対象にしてプロセスをkill
pkill -f 'go run'
# またはプロジェクト名などをヒントに:
# pkill -f 'your-server-binary-name'

echo "Go server process killed"
