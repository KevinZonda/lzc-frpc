build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./main.go

deploy: build
	main ./main.go
	npx lzc-cli project deploy
	npx lzc-cli project info

ssh:
	npx lzc-cli project exec /bin/sh

sync:
	npx lzc-cli project sync --watch