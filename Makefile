deploy:
	npx lzc-cli project deploy
	npx lzc-cli project info

ssh:
	npx lzc-cli project exec /bin/sh

sync:
	npx lzc-cli project sync --watch