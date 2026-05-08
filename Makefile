build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./main.go

build-lzc: build
	npx lzc-cli project build

build-release: build
	npx lzc-cli project build --release

deploy: build
	npx lzc-cli project deploy
	npx lzc-cli project info

deploy-release: build
	npx lzc-cli project deploy --release
	npx lzc-cli project info --release

ssh:
	npx lzc-cli project exec /bin/sh

sync:
	npx lzc-cli project sync --watch

clean:
	rm *.lpk | true
	rm -rf .lzc-cli-build* | true
