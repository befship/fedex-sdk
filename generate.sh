#!/usr/bin/env bash
# FedEx SDK generation pipeline:
#   Postman collections → postman2openapi → merge → patch → oapi-codegen
#
# Official Postman/OpenAPI are NEVER committed — only under .cache/.
#
# Modes:
#   1) Full Postman pipeline (preferred for updates):
#        POSTMAN_DIR=/path/to/collections make generate
#      Expects fedex-rate.json, fedex-ship.json, fedex-track.json (see scripts/manifest.yaml).
#   2) Direct OpenAPI (bootstrap / override):
#        SPEC_LOCAL=/path/to/openapi.json make generate
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

OAPI_CODEGEN_VERSION="${OAPI_CODEGEN_VERSION:-v2.5.0}"
export PATH="$(go env GOPATH)/bin:${PATH}"

if ! command -v oapi-codegen >/dev/null 2>&1; then
  echo "installing oapi-codegen@${OAPI_CODEGEN_VERSION} ..."
  go install "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}"
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 required" >&2
  exit 1
fi
python3 -c "import yaml" 2>/dev/null || {
  echo "PyYAML required: pip3 install pyyaml" >&2
  exit 1
}

MANIFEST="${ROOT}/scripts/manifest.yaml"
MERGE="${ROOT}/scripts/merge_oas.py"
PREPROCESS="${ROOT}/scripts/preprocess.py"
PATCH_DIR="${ROOT}/scripts/patches"
POSTMAN_CACHE="${ROOT}/.cache/postman"
SPEC_CACHE="${ROOT}/.cache/specs"
OUT_DIR="${ROOT}/internal/generated/fedexapi"
TMPDIR="$(mktemp -d "${TMPDIR:-/tmp}/fedex-sdk-gen.XXXXXX")"
trap 'rm -rf "$TMPDIR"' EXIT

mkdir -p "$POSTMAN_CACHE" "$SPEC_CACHE" "$OUT_DIR"

PKG="$(python3 - "$MANIFEST" <<'PY'
import sys, yaml
print(yaml.safe_load(open(sys.argv[1], encoding="utf-8"))["package"])
PY
)"
SERVER_SANDBOX="$(python3 - "$MANIFEST" <<'PY'
import sys, yaml
print(yaml.safe_load(open(sys.argv[1], encoding="utf-8"))["server_sandbox"])
PY
)"
SERVER_PROD="$(python3 - "$MANIFEST" <<'PY'
import sys, yaml
print(yaml.safe_load(open(sys.argv[1], encoding="utf-8"))["server_prod"])
PY
)"

MERGED="${SPEC_CACHE}/openapi.json"

build_from_postman() {
  local dir="$1"
  if ! command -v postman2openapi >/dev/null 2>&1; then
    echo "postman2openapi required: npm install -g postman2openapi" >&2
    return 1
  fi
  local yamls=()
  while IFS='|' read -r id file; do
    local src="${dir}/${file}"
    if [[ ! -f "$src" ]]; then
      echo "missing Postman collection: ${src}" >&2
      return 1
    fi
    cp "$src" "${POSTMAN_CACHE}/${file}"
    local out="${TMPDIR}/${id}.yaml"
    echo "→ postman2openapi ${file}"
    postman2openapi "$src" >"$out"
    yamls+=("$out")
  done < <(python3 - "$MANIFEST" <<'PY'
import sys, yaml
doc = yaml.safe_load(open(sys.argv[1], encoding="utf-8"))
for c in doc.get("collections") or []:
    print(f'{c["id"]}|{c["file"]}')
PY
)
  python3 "$MERGE" \
    --out "${TMPDIR}/merged.yaml" \
    --server-sandbox "$SERVER_SANDBOX" \
    --server-prod "$SERVER_PROD" \
    "${yamls[@]}"
  cp "${TMPDIR}/merged.yaml" "$MERGED"
}

if [[ -n "${SPEC_LOCAL:-}" && -f "${SPEC_LOCAL}" ]]; then
  echo "· local OpenAPI ${SPEC_LOCAL}"
  cp "${SPEC_LOCAL}" "$MERGED"
elif [[ -n "${POSTMAN_DIR:-}" ]]; then
  build_from_postman "$POSTMAN_DIR"
elif [[ -f "$MERGED" && "${FORCE_FETCH:-}" != "1" ]]; then
  echo "· cache hit ${MERGED}"
else
  cat >&2 <<'MSG'
No OpenAPI input.

Provide one of:
  SPEC_LOCAL=/path/to/openapi.json make generate
  POSTMAN_DIR=/path/to/postman-collections make generate

Download Rate/Ship/Track Postman collections from:
  https://developer.fedex.com/api/en-us/catalog.html
MSG
  exit 1
fi

CLEANED="${TMPDIR}/fedex-clean.json"
python3 "$PREPROCESS" \
  --in "$MERGED" \
  --out "$CLEANED" \
  --server-sandbox "$SERVER_SANDBOX" \
  --server-prod "$SERVER_PROD" \
  --patch-dir "$PATCH_DIR"

echo "→ generate ${PKG}"
SIZE=$(wc -c <"$CLEANED" | tr -d ' ')
if [[ "$SIZE" -gt 800000 ]]; then
  oapi-codegen -generate types -package "$PKG" -o "${OUT_DIR}/types.gen.go" "$CLEANED"
  oapi-codegen -generate client -package "$PKG" -o "${OUT_DIR}/client.gen.go" "$CLEANED"
else
  rm -f "${OUT_DIR}/types.gen.go"
  oapi-codegen -generate types,client -package "$PKG" -o "${OUT_DIR}/client.gen.go" "$CLEANED"
fi

if command -v gofmt >/dev/null 2>&1; then
  gofmt -w "${OUT_DIR}"/*.go
fi

echo "generate done."
