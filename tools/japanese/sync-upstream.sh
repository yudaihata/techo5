#!/usr/bin/env bash
# Prepare a reviewable merge of the latest stable Show release. Never modify a device or force-push.
set -euo pipefail
cd "$(dirname "$0")/../.."
base=codex/japanese
repo=$(gh repo view "$(git remote get-url origin)" --json nameWithOwner --jq .nameWithOwner)
if [[ "$repo" == HuskerMinion/* ]]; then
  echo 'Run this in your fork, not the official repository.' >&2; exit 1
fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo 'Commit or stash local changes first.' >&2; exit 1
fi
tag=$(gh api repos/HuskerMinion/techo5/releases/latest --jq .tag_name)
[[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Unexpected Show release tag' >&2; exit 1; }
git fetch origin "$base"
git fetch https://github.com/HuskerMinion/techo5.git "refs/tags/$tag"
release=$(git rev-parse FETCH_HEAD)
if git merge-base --is-ancestor "$release" "origin/$base"; then
  echo "$tag is already included."; exit 0
fi
branch="codex/upstream-$tag"
if gh pr list --repo "$repo" --head "$branch" --state open --json number --jq '.[0].number' | grep -q '[0-9]'; then
  echo "An update PR for $tag is already open."; exit 0
fi
if git show-ref --verify --quiet "refs/heads/$branch"; then
  echo "$branch already exists locally; inspect it before retrying." >&2; exit 1
fi
git switch -c "$branch" "origin/$base"
if ! git merge --no-edit "$release"; then
  echo 'Upstream changed overlapping code/history. Resolve this merge manually; nothing was pushed.' >&2
  exit 1
fi
printf '%s\n' "$tag" > tools/japanese/upstream-version.txt
git add tools/japanese/upstream-version.txt
if ! git diff --cached --quiet; then
  git commit -m "fork: track upstream Show release $tag"
fi
git push -u origin "$branch"
body=$(mktemp)
trap 'rm -f "$body"' EXIT
cat > "$body" <<EOF
Merge official Show release $tag into the Japanese branch, keeping the local changes as separate commits.

Review the merge and the Japanese fork checks before merging. No automatic merge or device installation is performed. The original release is https://github.com/HuskerMinion/techo5/releases/tag/$tag.
EOF
gh pr create --repo "$repo" --base "$base" --head "$branch" --title "upstream: merge $tag" --body-file "$body"
# GITHUB_TOKEN pushes do not trigger push/PR workflows; explicit dispatch does.
gh workflow run japanese-ci.yml --repo "$repo" --ref "$branch"
