import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import path from 'node:path';

const require = createRequire(import.meta.url);

const projectRoot = process.cwd();
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
