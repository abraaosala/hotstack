#!/usr/bin/env node

'use strict';

const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, 'bin');
const binPath = path.join(binDir, 'hotstack' + (process.platform === 'win32' ? '.exe' : ''));

if (fs.existsSync(binPath)) {
  fs.unlinkSync(binPath);
  console.log('[hotstack] Binário removido');
}
