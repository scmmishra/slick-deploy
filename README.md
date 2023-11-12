# Slick Deploy

Slick Deploy is a command-line tool designed to facilitate zero-downtime deployment for applications running in Docker containers, with Caddy as the reverse proxy.

## Features

- Zero-downtime deployment: Update your running application without interrupting service.
- Easy configuration: Use simple YAML files to manage deployment settings.
- Health checks: Ensure your application is running correctly before switching over.
- Rollback capability: Quickly revert to the previous version if something goes wrong.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Before you begin, ensure you have the following installed:

- Docker
- Caddy
- Go (1.15 or later)

### Installing

To start using Slick Deploy, install it right from the source:

```bash
git clone https://github.com/scmmishra/slick-deploy.git
cd slick-deploy
go build -o slick ./cmd/slick
```

You can now move the slick binary to a directory in your PATH to make it globally accessible.

### Usage

To deploy an application:

```bash
slick deploy --config path/to/your/config.yaml
```

To check the status of your deployment:

```bash
slick status
```

See `slick --help` for more information on commands and flags.

### Configuration

Create a config.yaml file with your deployment settings. Here's an example configuration:

```yaml
deployment:
  image_name: "ghcr.io/usememos/memos"
  container_port: 5230
  port_range:
    start: 3000
    end: 4000

caddy:
  admin_api: "http://localhost:2019"
  servers:
    - name: "*.memos.pages"
      reverse_proxy:
        - path: "/"
          to: ""

    - name: "api.memos.app"
      reverse_proxy:
        - path: "/api/*"
          to: "/api/*"

    - name: "dash.memos.app"
      reverse_proxy:
        - path: "/"
          to: "/dashboard/*"
        - path: "/preview"
          to: "/dashboard/preview"

health_check:
  endpoint: "/health"
  timeout_seconds: 5
```

### Development

To run the build with air use the following command

```bash
air --build.cmd "go build -o bin/slick cmd/slick/main.go" --build.bin ""
```

Add an alias to your .bashrc or .zshrc file to make it easier to run the slick command

```bash
alias slickdev="~/scmmishra/slick-deploy/bin/slick"
```

> Note, we use `slickdev` instead of `slick` to avoid conflicts with the global slick binary.
