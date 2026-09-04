#!/usr/bin/env node

'use strict';

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

const platform = os.platform();
const arch = os.arch();

const platforms = {
  'darwin-arm64': 'hotstack-darwin-arm64',
  'darwin-x64': 'hotstack-darwin-amd64',
  'linux-arm64': 'hotstack-linux-arm64',
  'linux-x64': 'hotstack-linux-amd64',
  'win32-x64': 'hotstack-windows-amd64',
};

const key = `${platform}-${arch}`;
const packageName = platforms[key];

if (!packageName) {
  console.warn(`[hotstack] Plataforma não suportada: ${platform}/${arch}`);
  console.warn('[hotstack] Hotstack pode não funcionar corretamente.');
  process.exit(0);
}

const binDir = path.join(__dirname, 'bin');
const binPath = path.join(binDir, 'hotstack' + (platform === 'win32' ? '.exe' : ''));

// Verificar se o binário já existe
if (fs.existsSync(binPath)) {
  console.log(`[hotstack] Binário já instalado: ${binPath}`);
  process.exit(0);
}

// Tentar encontrar o binário no node_modules do pacote da plataforma
const platformPackage = path.join(__dirname, '..', '..', packageName);
if (fs.existsSync(platformPackage)) {
  const platformBin = path.join(platformPackage, 'bin', 'hotstack' + (platform === 'win32' ? '.exe' : ''));
  if (fs.existsSync(platformBin)) {
    fs.mkdirSync(binDir, { recursive: true });
    fs.copyFileSync(platformBin, binPath);
    if (platform !== 'win32') {
      fs.chmodSync(binPath, 0o755);
    }
    console.log(`[hotstack] Instalado de ${packageName}`);
    process.exit(0);
  }
}

// Se não encontrou, avisar
console.warn(`[hotstack] Binário não encontrado para ${packageName}`);
console.warn('[hotstack] Execute: hotstack update');
