#!/usr/bin/env bash
# Update gitea/ - the parts of Gitea this plugin builds into Caddy - from upstream.
#
#   releases/sync-gitea.sh [gitea-ref]
#
#   gitea-ref  upstream go-gitea/gitea branch, tag or commit; empty = main
#
# 1. Upstream Gitea with its whole history, plus wekan/gitea-patches (its Finnish
#    translation and its patches), as commits - by gitea-patches' own
#    releases/apply-patches.sh, so both repositories make the same Gitea.
# 2. This repository's dist/gitea/*.patch on top, as commits (git am --3way).
# 3. Gitea's frontend is built (make deps-frontend frontend): its public/assets
#    are among the embedded files.
# 4. releases/copy-gitea.py replaces gitea/ with the needed packages; go.mod gets
#    Gitea's dependency versions; then go mod tidy, go vet and go build.
#
# It changes the working tree only - committing is the maintainer's step - and
# writes gitea/SOURCE.md with the exact commits it was made from.
#
# Environment:
#   GITEA_PATCHES   a wekan/gitea-patches checkout or URL (default: ../gitea-patches
#                   when it exists - WeKan's .tools/w/ - else its GitHub URL)
#   UPSTREAM_URL    where upstream Gitea is cloned from (passed to gitea-patches)
#   SYNC_WORK       work directory to keep (default: a temporary one, removed)
# Needs git, Go, Node.js and pnpm (Gitea's package.json names the pnpm version).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REF="${1:-}"
if [ -z "${GITEA_PATCHES:-}" ]; then
  if [ -d "$ROOT/../gitea-patches/releases" ]; then GITEA_PATCHES="$ROOT/../gitea-patches"
  else GITEA_PATCHES=https://github.com/wekan/gitea-patches.git; fi
fi
WORK="${SYNC_WORK:-$(mktemp -d "${TMPDIR:-/tmp}/caddy-gitea-sync.XXXXXX")}"
[ -n "${SYNC_WORK:-}" ] || trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK"

if [ -d "$GITEA_PATCHES/releases" ]; then
  GP="$(cd "$GITEA_PATCHES" && pwd)"
else
  git clone --quiet "$GITEA_PATCHES" "$WORK/gitea-patches"
  GP="$WORK/gitea-patches"
fi
GP_COMMIT="$(git -C "$GP" rev-parse HEAD 2>/dev/null || echo unknown)"

echo "== 1. upstream Gitea + gitea-patches"
rm -rf "$WORK/gitea"
bash "$GP/releases/apply-patches.sh" "$GP" "$WORK/gitea" "$REF" | tee "$WORK/apply.txt"
UPSTREAM_COMMIT="$(sed -n 's/^UPSTREAM_COMMIT=//p' "$WORK/apply.txt")"
UPSTREAM_VERSION="$(sed -n 's/^UPSTREAM_VERSION=//p' "$WORK/apply.txt")"
SOURCE_REF="$(sed -n 's/^SOURCE_REF=//p' "$WORK/apply.txt")"

echo "== 2. this plugin's patches"
cd "$WORK/gitea"
export GIT_COMMITTER_NAME="Lauri Ojansivu" GIT_COMMITTER_EMAIL="x@xet7.org"
export GIT_COMMITTER_DATE="$(git log -1 --format=%cI "$UPSTREAM_COMMIT")"
for p in "$ROOT"/dist/gitea/*.patch; do
  [ -e "$p" ] || continue
  n="$(basename "$p")"
  ( cd "$ROOT/dist/gitea" && sha256sum --quiet -c "${n%.patch}.sha256sum" )
  echo "Applying dist/gitea/${n}"
  git am --quiet --3way --committer-date-is-author-date "$p" || {
    git am --abort || true
    echo "ERROR: dist/gitea/${n} does not apply to Gitea ${SOURCE_REF} (${UPSTREAM_COMMIT}). Re-port it." >&2
    exit 1
  }
done
GITEA_COMMIT="$(git rev-parse HEAD)"
git log --oneline --no-decorate "${UPSTREAM_COMMIT}..HEAD"

echo "== 3. Gitea's frontend"
make deps-frontend frontend

echo "== 4. copy the needed parts"
cd "$ROOT"
python3 -I releases/copy-gitea.py "$WORK/gitea" "$ROOT"
# Gitea's dependency versions are the floor: the copied code was tested with them.
go mod edit -go="$(sed -n 's/^go //p' "$WORK/gitea/go.mod")"
toolchain="$(sed -n 's/^toolchain //p' "$WORK/gitea/go.mod")"
[ -z "$toolchain" ] || go mod edit -toolchain="$toolchain"
( cd "$WORK/gitea" && go list -m -f '{{if not .Main}}{{.Path}}@{{.Version}}{{end}}' all ) | while read -r req; do
  [ -n "$req" ] && [ "${req##*@}" != "" ] && echo "-require=$req"
done | xargs -r go mod edit
GOFLAGS=-mod=mod go mod tidy
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go build ./...

cat > gitea/SOURCE.md <<EOF
# gitea/ - the parts of Gitea this plugin builds into Caddy

Copied by \`releases/sync-gitea.sh\` - do not edit here: change upstream Gitea,
wekan/gitea-patches, or this repository's \`dist/gitea/\`, and sync again.

| | |
|---|---|
| Upstream | [go-gitea/gitea](https://github.com/go-gitea/gitea) \`${SOURCE_REF}\`, version ${UPSTREAM_VERSION}, commit [\`${UPSTREAM_COMMIT}\`](https://github.com/go-gitea/gitea/commit/${UPSTREAM_COMMIT}) |
| wekan/gitea-patches | commit \`${GP_COMMIT}\` |
| This repository's patches | $(ls "$ROOT"/dist/gitea/*.patch 2>/dev/null | xargs -r -n1 basename | sed 's/^/`/;s/$/`/' | paste -sd, - | sed 's/,/, /g') |
| Resulting Gitea commit | \`${GITEA_COMMIT}\` |

Only the packages the plugin needs are here - \`go list -deps\` of what it imports,
for every operating system Weddy is built for - with Gitea's assets embedded from
\`assets/\` directories. Gitea is MIT licensed: [LICENSE](LICENSE).
EOF
echo "gitea/ now holds Gitea ${UPSTREAM_VERSION} (${UPSTREAM_COMMIT:0:10}) with $(ls dist/gitea/*.patch 2>/dev/null | wc -l) plugin patch(es)."
