#!/bin/sh
set -e

REPO="abraaosala/hotstack"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

detect_platform() {
  os=$(uname -s)
  arch=$(uname -m)

  case "$os" in
    Linux)  platform="linux" ;;
    Darwin) platform="darwin" ;;
    *)      echo "SO não suportado: $os"; exit 1 ;;
  esac

  case "$arch" in
    x86_64|amd64)  platform="${platform}-amd64" ;;
    aarch64|arm64) platform="${platform}-arm64" ;;
    *)             echo "Arquitetura não suportada: $arch"; exit 1 ;;
  esac

  echo "$platform"
}

get_latest_version() {
  curl -sL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

main() {
  echo "HotStack - Instalador"
  echo "====================="
  echo ""

  platform=$(detect_platform)
  version=$(get_latest_version)

  if [ -z "$version" ]; then
    echo "Erro: não foi possível detetar a versão mais recente"
    exit 1
  fi

  echo "Plataforma: $platform"
  echo "Versão:     $version"
  echo ""

  filename="hotstack-${platform}${version:+-#}"
  url="https://github.com/$REPO/releases/download/${version}/hotstack-${platform}"

  echo "A descargar de: $url"

  mkdir -p "$INSTALL_DIR"

  if curl -sL -o "$INSTALL_DIR/hotstack" "$url"; then
    chmod +x "$INSTALL_DIR/hotstack"
    echo ""
    echo "✓ HotStack instalado em: $INSTALL_DIR/hotstack"
    echo ""
    echo "Adiciona ao PATH (se necessário):"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  else
    echo "Erro ao descargar hotstack"
    exit 1
  fi
}

main "$@"
