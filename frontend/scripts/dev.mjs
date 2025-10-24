import { existsSync } from 'node:fs';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';

const npmCommand = process.platform === 'win32' ? 'npm.cmd' : 'npm';
const projectRoot = process.cwd();
const viteBinary = join(
  projectRoot,
  'node_modules',
  '.bin',
  process.platform === 'win32' ? 'vite.cmd' : 'vite'
);

if (!existsSync(viteBinary)) {
  console.log('Installing dependencies with `npm install`...');
  const installResult = spawnSync(npmCommand, ['install'], {
    cwd: projectRoot,
    stdio: 'inherit'
  });

  if (installResult.status !== 0) {
    console.error('Failed to install dependencies. Aborting.');
    process.exit(installResult.status ?? 1);
  }
}

const execResult = spawnSync(npmCommand, ['exec', 'vite', ...process.argv.slice(2)], {
  cwd: projectRoot,
  stdio: 'inherit'
});

if (execResult.status !== 0) {
  process.exit(execResult.status ?? 1);
}
