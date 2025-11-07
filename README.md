# ⚡ Blast

Zero-config local HTTPS reverse proxy for development.

## What is Blast?

Blast maps your local development servers to trusted HTTPS domains instantly. No configuration, no manual certificate management, no hosts file editing.

```bash
sudo blast start 3000 myapp
# https://myapp.blast is now live and trusted
```

## Features

- **Zero Configuration** - Works out of the box
- **Automatic HTTPS** - Self-signed certificates trusted by your system
- **Single Binary** - No dependencies
- **Cross-Platform** - macOS, Linux, Windows

## Installation

```bash
go install github.com/doganarif/blast/cmd/blast@latest
```

Or build from source:

```bash
git clone https://github.com/doganarif/blast.git
cd blast
go build -o blast ./cmd/blast
sudo mv blast /usr/local/bin/
```

## Usage

### Start a proxy

```bash
sudo blast start 3000 api
```

This automatically:
- Generates CA certificate (first run only)
- Creates SSL certificate for `api.blast`
- Adds `127.0.0.1 api.blast` to `/etc/hosts`
- Starts background daemon
- Routes `https://api.blast` → `http://localhost:3000`

Visit `https://api.blast` in your browser. It just works.

### List proxies

```bash
blast list
```

### Stop a proxy

```bash
sudo blast stop api
```

### Firefox Users

Firefox uses its own certificate store. To enable HTTPS in Firefox:

```bash
blast ca-path
```

Follow the instructions to import the CA certificate into Firefox.

## How It Works

1. First run generates a root CA and installs it in your system trust store
2. For each domain, Blast generates a certificate signed by the CA
3. Background daemon listens on port 443 and reverse-proxies to your local ports
4. Hosts file entries route `*.blast` domains to `127.0.0.1`

## Requirements

- Root/Administrator privileges (for port 443, CA installation, hosts file)
- Go 1.21+ (for building from source)

## Platform Support

- macOS: Uses `security` keychain
- Linux: Uses `update-ca-certificates` or `update-ca-trust`
- Windows: Uses `certutil`

## Architecture

```
blast/
├── cmd/blast/          # CLI entry point
└── internal/
    ├── ca/             # Certificate authority
    ├── cert/           # SSL certificate generation
    ├── config/         # Configuration persistence
    ├── daemon/         # Background process management
    ├── hosts/          # /etc/hosts management
    ├── proxy/          # HTTPS reverse proxy
    └── system/         # Platform-specific operations
```

## Configuration

- Config: `~/.config/blast/config.json`
- CA certificates: `~/.config/blast/ca/`
- Daemon logs: `~/.config/blast/daemon.log`
- PID file: `~/.config/blast/daemon.pid`

## License

MIT - see [LICENSE](LICENSE)

## Contributing

Issues and pull requests welcome.
