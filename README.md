# Lazycat Golang Todo App

## First Deploy
```bash
lzc-cli project deploy
lzc-cli project info
```

By default, project commands use `lzc-build.dev.yml` when it exists.
Each command prints the active `Build config`.
Use `--release` if you want to operate on `lzc-build.yml`.

Open the app first.
In dev mode, the app will not auto-start the Golang service.
If the backend service is not ready yet, the app page will show the expected port and what to do next.

## Recommended Backend Dev Loop
```bash
lzc-cli project sync --watch
lzc-cli project exec /bin/sh
/app/run.sh
```

You can also build locally and use `lzc-cli project cp` instead of `project sync --watch`.

## Troubleshooting
```bash
lzc-cli project log -f
lzc-cli project exec /bin/sh
```

## API
```text
GET /api/health
GET /api/todos
POST /api/todos
PUT /api/todos/{id}/toggle
DELETE /api/todos/{id}
```

## Data Path
```text
/lzcapp/var/todos.json
```
