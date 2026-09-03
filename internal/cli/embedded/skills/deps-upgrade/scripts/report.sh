#!/usr/bin/env bash
# Report outdated direct dependencies
# Detect the manifest and run the appropriate command.
set -e

if [ -f composer.json ]; then
  composer outdated --direct
elif [ -f package.json ]; then
  npm outdated
elif [ -f Cargo.toml ]; then
  cargo upgrade --dry-run --incompatible
elif [ -f pyproject.toml ]; then
  uv tree --outdated
elif [ -f go.mod ]; then
  go list -m -u all
else
  echo "Nenhum manifest encontrado (composer.json, package.json, Cargo.toml, pyproject.toml, go.mod)"
  exit 1
fi
