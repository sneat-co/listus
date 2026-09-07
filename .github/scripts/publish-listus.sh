#!/usr/bin/env bash
# Runs Nx release for the listus npm packages and lands the resulting
# version-bump commit on a plain (non-protected) branch instead of pushing
# straight to main. Main requires status checks, so a direct push from this
# job's token is rejected with GH006. The `open-release-pr` job in
# publish.yml opens the actual pull request afterwards using an org PAT,
# because a push/PR made with the default GITHUB_TOKEN does not trigger the
# checks that PR needs to satisfy those requirements.
#
# Idempotent: if Nx's publish step fails only because the version it
# resolved is already on npm (e.g. a rerun after a prior push failure like
# run 34127816455), that failure is downgraded to a notice instead of
# failing the job.
set -uo pipefail

before_head="$(git rev-parse HEAD)"

output="$(pnpm exec nx release --yes 2>&1)"
status=$?
printf '%s\n' "${output}"

if [[ ${status} -ne 0 ]]; then
  if printf '%s\n' "${output}" | grep -qiE 'you cannot publish over the previously published version|EPUBLISHCONFLICT'; then
    echo "::notice::Resolved listus version is already published to npm; treating publish as already done."
  else
    exit "${status}"
  fi
fi

after_head="$(git rev-parse HEAD)"
if [[ "${before_head}" == "${after_head}" ]]; then
  echo "::notice::Nx release made no local commit (no release-worthy changes); nothing to land on main."
  exit 0
fi

# Tags are not protected refs, so push the release tag directly.
tag="$(git describe --tags --exact-match HEAD)"
git push origin "refs/tags/${tag}"

# Land the version-bump commit via a plain branch; open-release-pr turns it
# into a pull request against protected main.
branch="release/listus-${GITHUB_RUN_ID:-local}"
git switch --create "${branch}"
git push --set-upstream origin "${branch}"
