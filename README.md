# hotreload

A lightweight CLI tool that watches your project files, rebuilds on change, and restarts your process automatically.

## Features

- **File watching** — detects changes across your entire project tree
- **Auto rebuild** — runs your build command before restarting
- **Graceful restart** — sends SIGINT to the running process, waits up to 5s, then force kills
- **Debounced triggers** — rapid saves only fire one reload
- **New directory detection** — directories created at runtime are automatically watched
- **Crash loop protection** — exponential backoff (2s → 4s → 8s, capped at 30s) if the process keeps crashing on startup
- **In-progress build cancellation** — a new file change discards any running build immediately

## Installation

**From source (Linux / Mac):**
```sh
make install
```

**Windows:**
```powershell
.\build.ps1
# then move bin\hotreload.exe somewhere on your PATH
```

## Usage

```sh
hotreload [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--root`, `-r` | `.` | Root directory to watch |
| `--build`, `-b` | _(none)_ | Build command to run before restarting (optional) |
| `--exec`, `-e` | _(required)_ | Command to run |
| `--debounce` | `500` | Debounce delay in milliseconds |
| `--verbose`, `-v` | `false` | Enable debug logging |

### Examples

**Go project** — rebuild and restart on every change:
```sh
hotreload --root . --build "go build -o ./bin/app ." --exec "./bin/app"
```

**Python / Node / any interpreter** — no build step needed:
```sh
hotreload --root . --exec "python main.py"
hotreload --root . --exec "node server.js"
```

**Watch a subdirectory only:**
```sh
hotreload --root ./src --build "go build -o ./bin/app ." --exec "./bin/app"
```

## Building

**Linux / Mac — using Make:**
```sh
make build        # build for current OS  →  bin/hotreload
make linux        # cross-compile         →  bin/hotreload-linux
make mac          # cross-compile         →  bin/hotreload-mac
```

**Windows — using PowerShell:**
```powershell
.\build.ps1       # build  →  bin\hotreload.exe
```

## Testing

```sh
make test         # run all unit tests
make test-race    # run with race detector
make cover        # generate coverage report
```

## Project Structure

```
hotreload/
├── cmd/hotreload/       # CLI entry point
├── internal/
│   ├── builder/         # runs the build command
│   ├── debouncer/       # collapses rapid file events into one trigger
│   ├── runner/          # starts, stops, and monitors the process
│   └── watcher/         # watches the filesystem for changes
└── testserver/          # sample Go HTTP server used for manual testing
```

## Ignored Paths

The watcher automatically skips these directories:

`.git` `.github` `node_modules` `vendor` `venv` `.venv` `dist` `build` `bin` `out` `.vscode` `.idea` `.cache` `.next` `.nuxt` `__pycache__`

And these file extensions:

`.log` `.tmp` `.swp` `.bak` `.exe` `.dll` `.so` `.o` `.a` `.zip` `.tar` `.gz`
