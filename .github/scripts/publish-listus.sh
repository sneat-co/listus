#!/usr/bin/env bash
# Runs Nx release for the listus npm packages and lands the resulting
# version-bump commit on a plain (non-protected) branch instead of pushing
# straight to main. Main requires status checks, so a direct push from this
# job's token is rejected with GH006. The `open-release-pr` job in
# publish.yml opens the actual pull request afterwards using an org PAT,
# because a push/PR made with the default GITHUB_TOKEN does not trigger the
# checks that PR needs to satisfy those requirements.
#
# Idempotent: if Nx fails only because the version it resolved was already
# released (npm already has it, or the local git tag for it already exists),
# that failure is downgraded to a notice instead of failing the job. Both
# cases show up here after a prior run's version-bump commit failed to reach
# main (e.g. run 34127816455): Nx still resolves "current version" from the
# nearest ancestor git tag, so it can recompute the same already-released
# version and either try to re-publish it (npm rejects that) or try to
# recreate its already-pushed release tag (git rejects that too, and it
# happens even earlier in Nx's pipeline, before publish is attempted).
set -uo pipefail

before_head="$(git rev-parse HEAD)"

# Nx resolves the current version from the nearest ancestor tag. A prior
# protected-main release can publish and tag successfully while its release
# commit is still waiting in a PR, leaving the newest tag on a sibling branch.
# In that state, infer the next patch from the public registry so the next
# verified main push cannot silently replay the already-published version.
release_args=(--yes)
published_version="$(npm view @sneat/extension-listus version 2>/dev/null || true)"
if [[ "${published_version}" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]] &&
   git rev-parse -q --verify "refs/tags/v${published_version}" >/dev/null &&
   ! git merge-base --is-ancestor "refs/tags/v${published_version}" HEAD; then
  next_patch="$((BASH_REMATCH[3] + 1))"
  release_args=("${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.${next_patch}" --yes)
  echo "::notice::Newest published Listus tag is not an ancestor of HEAD; releasing ${release_args[0]} explicitly."
fi

output="$(pnpm exec nx release "${release_args[@]}" 2>&1)"
status=$?
printf '%s\n' "${output}"

if [[ ${status} -ne 0 ]]; then
  if printf '%s\n' "${output}" | grep -qiE "you cannot publish over the previously published version|EPUBLISHCONFLICT|tag '[^']*' already exists"; then
    echo "::notice::Resolved listus version was already released (published to npm and/or already tagged); treating this run as already done."
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
