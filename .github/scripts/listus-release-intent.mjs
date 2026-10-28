import { appendFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';

const releaseSubject = /^chore\(release\): publish (?:listus )?\d+\.\d+\.\d+(?: \(#\d+\))?$/;

export function shouldPublish({ eventName, message = '', repository }) {
  if (eventName === 'workflow_dispatch') return true;
  const [subject, ...body] = message.trim().split(/\r?\n/);
  if (releaseSubject.test(subject)) return false;
  const merge = /^Merge pull request #\d+ from (\S+)\/release\/listus-\d+$/.exec(subject);
  const owner = repository?.split('/')[0];
  return !(merge?.[1] === owner && releaseSubject.test(body.join('\n').trim()));
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const publish = shouldPublish({
    eventName: process.env.GITHUB_EVENT_NAME,
    message: process.env.LISTUS_HEAD_MESSAGE,
    repository: process.env.GITHUB_REPOSITORY,
  });
  appendFileSync(process.env.GITHUB_OUTPUT, `publish=${publish}\n`);
}
