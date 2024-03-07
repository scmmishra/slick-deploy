<div align="center">
<br>
<br>
<p>
  <img src="./.github/logo-light.svg" alt="SlickDeploy" width="300">
</p>
<br>
<br>

</div>

# Slick Deploy

Slick Deploy is a command-line tool designed to facilitate zero-downtime deployment for applications running in Docker containers, with Caddy as the reverse proxy.

> [!WARNING]
> Slick is not ready for production, but you can give it a shot if you're feeling adventurous

### Installing

To start using Slick Deploy, install it using this one-line command:

```bash
curl -fsSL https://raw.githubusercontent.com/scmmishra/slick-deploy/main/scripts/install.sh | bash
```

It will install the latest version of slick in /usr/local/bin/slick. You can use the same script to update the CLI.

To install from source manually, run the following commands:

```bash
git clone https://github.com/scmmishra/slick-deploy.git
cd slick-deploy
make install
```

## Features

- Zero-downtime deployment: Update your running application without interrupting service.
- Easy configuration: Use simple YAML files to manage deployment settings.
- Health checks: Ensure your application is running correctly before switching over.
- Rollback capability: Quickly revert to the previous version if something goes wrong.

## Why?

Just for fun, I couldn't find a tool that was minimal and had near zero-downtime deploys. All I was looking for something that worked as a slim layer between me and the tools I use to run apps on a VM. So I built it.
I also wanted to learn Go for a while, and Go is simply amazing for building CLI tools.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Before you begin, ensure you have the following installed:

- Docker
- Caddy
- Go (1.15 or later)

### Usage

To deploy an application:

```bash
slick deploy --config path/to/your/config.yaml --env path/to/your/.env
```

To check the status of your deployment:

```bash
slick status
```

To check logs for your deployment:

```bash
slick logs
```

See `slick --help` for more information on commands and flags.

### Configuration

Create a config.yaml file with your deployment settings. Here's an example configuration:

```yaml
app:
  name: "memos"
  image: "ghcr.io/usememos/memos"
  container_port: 5230
  registry:
    username: "<username>"
    password: SLICK_REGISTRY_PASSWORD
  env:
    - AWS_S3_ACCESS_KEY_ID
    - AWS_S3_BUCKET_NAME
    - AWS_S3_CUSTOM_DOMAIN
    - AWS_S3_ENDPOINT_URL
    - AWS_S3_REGION_NAME
    - AWS_S3_SECRET_ACCESS_KEY
  port_range:
    start: 8000
    end: 9000

caddy:
  admin_api: "http://localhost:2019"
  rules:
    - match: "*.pages.dev"
      reverse_proxy:
        - path: ""
          to: "http://localhost:{port}"

    - match: "localhost"
      reverse_proxy:
        - path: ""
          to: "localhost:{port}"
        - path: "/api/*"
          to: "localhost:{port}/internal/api/*"

health_check:
  endpoint: "/health"
  timeout_seconds: 5
```

### Managing environment variables

You can point to an `.env` file to load environment variables from. This is useful for storing sensitive information like passwords and API keys.

```bash
slick deploy --config path/to/your/config.yaml --env path/to/your/.env
```

However, it is best to use a tool like [Phase](https://phase.dev) to manage your environment variables. Phase allows you to store your environment variables in a secure, encrypted vault, and then inject them into your application at runtime.

```bash
phase run slick deploy
```

Read more about Phase CLI [here](https://docs.phase.dev/cli/commands).

### Development

To run the build with air use the following command

```bash
air --build.cmd "go build -o bin/slick cmd/slick/main.go" --build.bin ""
```

Add an alias to your .bashrc or .zshrc file to make it easier to run the slick command

```bash
alias slickdev="<path-to-project>/bin/slick"
```

> Note, we use `slickdev` instead of `slick` to avoid conflicts with the global slick binary.

#### Testing

To run the tests, use the following command

```bash
go test ./... -coverprofile=coverage.out
```

This will also generate the coverage report, which can be viewed by running

```bash
go tool cover -html=coverage.out
```
