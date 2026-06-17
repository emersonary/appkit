import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const tenantsRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const tenantBlockRoot = resolve(tenantsRoot, '..');
const viaJeriRoot = resolve(tenantBlockRoot, '../../..');
const protoDir = join(tenantBlockRoot, 'proto');
const outDir = join(tenantsRoot, 'src', 'gen');

function plugin(name) {
  const win = process.platform === 'win32';
  const candidates = [
    join(tenantsRoot, 'node_modules', '.bin'),
    join(viaJeriRoot, 'node_modules', '.bin'),
    join(viaJeriRoot, 'appkit/blocks/account/web/node_modules', '.bin'),
  ];
  for (const bin of candidates) {
    const cmd = join(bin, win ? `${name}.cmd` : name);
    const sh = join(bin, name);
    if (existsSync(cmd)) return cmd;
    if (existsSync(sh)) return sh;
  }
  return join(candidates[0], win ? `${name}.cmd` : name);
}

const protoFile = join(protoDir, 'v1', 'tenant.proto');

mkdirSync(outDir, { recursive: true });

execFileSync(
  'protoc',
  [
    '-I',
    protoDir,
    `--plugin=protoc-gen-es=${plugin('protoc-gen-es')}`,
    `--plugin=protoc-gen-connect-es=${plugin('protoc-gen-connect-es')}`,
    `--es_out=${outDir}`,
    '--es_opt=target=ts',
    `--connect-es_out=${outDir}`,
    '--connect-es_opt=target=ts',
    protoFile,
  ],
  { stdio: 'inherit' },
);

console.log('Generated tenant TypeScript into appkit/blocks/tenant/web/src/gen/');
