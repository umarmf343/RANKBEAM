import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);

// Resolve the project root relative to this script to avoid issues where
// process.cwd() is not the repository root (for example when npm is executed
// from a different directory on Windows).
const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const projectRoot = path.resolve(scriptDir, '..');

function resolveViteBin() {
  const vitePackageRoot = require.resolve('vite/package.json', { paths: [projectRoot] });
  const viteDir = path.dirname(vitePackageRoot);
  return path.join(viteDir, 'bin', 'vite.js');
}

function getNpmCommand() {
  const npmExecPath = process.env.npm_execpath;

  if (npmExecPath) {
    return {
      command: process.execPath,
      args: [npmExecPath, 'install']
    };
  }

  const npmCommand = process.platform === 'win32' ? 'npm.cmd' : 'npm';
  return {
    command: npmCommand,
    args: ['install']
  };
}

function installDependencies() {
  const { command, args } = getNpmCommand();
  const installResult = spawnSync(command, args, {
    cwd: projectRoot,
    stdio: 'inherit',
    shell: command === 'npm.cmd'
  });

  if (installResult.error) {
    console.error('Failed to run `npm install` automatically.', installResult.error);
    process.exit(1);
  }

  if (installResult.status !== 0) {
    process.exit(installResult.status ?? 1);
  }
}

let viteBinPath;
try {
  viteBinPath = resolveViteBin();
} catch (error) {
  console.warn('Dependencies are missing. Installing them before starting the dev server...');
  installDependencies();
  try {
    viteBinPath = resolveViteBin();
  } catch (secondError) {
    console.error('Unable to resolve Vite even after installing dependencies.');
    if (secondError instanceof Error) {
      console.error(secondError.message);
    }
    process.exit(1);
  }
}

const execResult = spawnSync(process.execPath, [viteBinPath, ...process.argv.slice(2)], {
  cwd: projectRoot,
  stdio: 'inherit'
});

if (execResult.status !== 0) {
  process.exit(execResult.status ?? 1);
}
