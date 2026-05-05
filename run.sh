#!/bin/sh
set -e

mirror_root="/lzcapp/cache/project-mirror"
if command -v go >/dev/null 2>&1 && [ -f "$mirror_root/main.go" ]; then
	cd "$mirror_root"
	exec ./app
fi

exec /app/app
