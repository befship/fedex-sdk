#!/usr/bin/env bash
# Compatibility entrypoint — patching is scripts/preprocess.py + scripts/patches/*.json
set -euo pipefail
echo "Use: make generate  (Postman→OAS→merge→patch→oapi-codegen)" >&2
echo "Add structural fixes under scripts/patches/*.json" >&2
exit 0
