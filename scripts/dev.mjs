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
let viteBinPath;
try {
  const vitePackageRoot = require.resolve('vite/package.json', { paths: [projectRoot] });
  const viteDir = path.dirname(vitePackageRoot);
  viteBinPath = path.join(viteDir, 'bin', 'vite.js');
} catch (error) {
  console.error(
    'Dependencies are missing. Please run `npm install` in the project root before starting the dev server.'
  );
  if (error instanceof Error) {
    console.error(error.message);
  }
  process.exit(1);
}

const execResult = spawnSync(process.execPath, [viteBinPath, ...process.argv.slice(2)], {
  cwd: projectRoot,
  stdio: 'inherit'
});

if (execResult.status !== 0) {
  process.exit(execResult.status ?? 1);
}
