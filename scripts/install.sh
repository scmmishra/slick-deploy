#!/bin/bash
# This script installs the slick deploy to /usr/local/bin/slick from Github release
args=("$@")
custom_version=${args[0]}

# Function to detect latest GitHub release
get_latest_release() {
  local repo=$1
  curl -s "https://api.github.com/repos/${repo}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

if [ -z "$custom_version" ]; then
  custom_version=$(get_latest_release "scmmishra/slick-deploy")
fi

# Function to detect platform, architecture, etc.
detect_platform() {
  OS=$(uname -s)
  ARCH=$(uname -m)

  case $OS in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported operating system: $OS"
    exit 1
    ;;
  esac

  case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
  esac

  echo "Detected platform: $OS $ARCH"
}

# Function to download file from GitHub release
download_from_github() {
  local repo=$1
  local release=$2
  local name=$3

  local release_without_prefix=${release#v}
  local filename=${name}_${release_without_prefix}_${OS}_${ARCH}.tar.gz
  # Construct download URL
  local download_url="https://github.com/${repo}/releases/download/${release}/${filename}"

  # Use curl to download the file quietly
  echo "Downloading ${name} ${release} from GitHub release"
  curl -sL -o "${filename}" "${download_url}"

  # Determine the binary directory
  local binary_dir=""
  if [ "$OS" == "linux" ] || [ "$OS" == "darwin" ]; then
    binary_dir="/usr/local/bin"
  fi
  echo "Installing ${name}"
  sudo tar -xzvf "${filename}" -C "${binary_dir}" > /dev/null

  # Make the binary executable
  sudo chmod +x "${binary_dir}/slick"

  # Cleanup
  rm "${filename}"

  echo "${name} installed successfully to ${binary_dir}/slick"
}

echo "⠠⠞⠑⠭⠞⠶⠠⠑⠙⠊⠞⠕⠗ Slick Deploy ⠠⠞⠑⠭⠞⠶⠠⠑⠙⠊⠞⠕⠗"
detect_platform
download_from_github "scmmishra/slick-deploy" $custom_version "slick-deploy"
