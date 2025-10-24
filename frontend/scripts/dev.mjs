import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

const npmCommand = process.platform === 'win32' ? 'npm.cmd' : 'npm';
const projectRoot = process.cwd();
try {
  require.resolve('vite/package.json', { paths: [projectRoot] });
} catch (error) {
  console.error(
    'Dependencies are missing. Please run `npm install` in the project root before starting the dev server.'
  );
  if (error instanceof Error) {
    console.error(error.message);
  }
  process.exit(1);
}

const execResult = spawnSync(npmCommand, ['exec', 'vite', ...process.argv.slice(2)], {
  cwd: projectRoot,
  stdio: 'inherit'
});

if (execResult.status !== 0) {
  process.exit(execResult.status ?? 1);
}
