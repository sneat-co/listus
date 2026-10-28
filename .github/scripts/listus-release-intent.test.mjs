import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import test from 'node:test';
import { shouldPublish } from './listus-release-intent.mjs';

const release = 'chore(release): publish listus 0.1.23';
const check = (message, eventName = 'push') => shouldPublish({ eventName, message, repository: 'sneat-co/listus' });

for (const message of [release, `${release} (#99)`, 'chore(release): publish 0.1.23', `Merge pull request #99 from sneat-co/release/listus-123\n\n${release}`]) {
  test(`skips generated bookkeeping: ${message.split('\n')[0]}`, () => assert.equal(check(message), false));
}

for (const message of [
  'feat(listus): add a useful feature',
  `fix(listus): preserve publication\n\nThe old message was ${release}.`,
  `Merge pull request #99 from sneat-co/fix-release\n\n${release}`,
  `Merge pull request #99 from sneat-co/release/listus-123\n\nfix: quoted ${release}`,
  `Merge pull request #99 from another-owner/release/listus-123\n\n${release}`,
]) {
  test(`publishes real change: ${message.split('\n')[0]}`, () => assert.equal(check(message), true));
}

test('explicit manual dispatch can publish even at a bookkeeping head', () => assert.equal(check(release, 'workflow_dispatch'), true));
test('missing push message does not suppress publication', () => assert.equal(check(undefined), true));
test('missing repository cannot turn a real fix into bookkeeping', () => {
  assert.equal(shouldPublish({ eventName: 'push', message: `fix: quote\n\n${release}` }), true);
});
test('CLI writes the GitHub job output from environment input', () => {
  const dir = mkdtempSync(join(tmpdir(), 'listus-release-intent-'));
  try {
    const output = join(dir, 'output');
    const result = spawnSync(process.execPath, [new URL('./listus-release-intent.mjs', import.meta.url).pathname], {
      env: { ...process.env, GITHUB_EVENT_NAME: 'push', LISTUS_HEAD_MESSAGE: release, GITHUB_REPOSITORY: 'sneat-co/listus', GITHUB_OUTPUT: output },
    });
    assert.equal(result.status, 0, result.stderr.toString());
    assert.equal(readFileSync(output, 'utf8'), 'publish=false\n');
  } finally { rmSync(dir, { recursive: true }); }
});
