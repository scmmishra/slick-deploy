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
  image_name: "your-app-image"
  container_base_name: "your-app-container"

caddy:
  admin_api: "http://localhost:2019"
  proxy_matcher: "example.com"

health_check:
  endpoint: "/health"
  timeout_seconds: 5
```
