#!/usr/bin/env node

'use strict';

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, '..', 'bin');
const binPath = path.join(binDir, 'hotstack' + (process.platform === 'win32' ? '.exe' : ''));

if (!fs.existsSync(binPath)) {
  console.error('[hotstack] Binário não encontrado. Execute: hotstack update');
  process.exit(1);
}

const args = process.argv.slice(2).map(arg => `"${arg}"`).join(' ');
const cmd = `"${binPath}" ${args}`;

try {
  const result = execSync(cmd, {
    stdio: 'inherit',
    cwd: process.cwd(),
  });
  process.exit(result.status || 0);
} catch (error) {
  process.exit(error.status || 1);
}
