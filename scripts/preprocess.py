#!/usr/bin/env python3
"""Patch FedEx OpenAPI derived from Postman (required flags, enums, error models).

This is intentionally conservative: Postman→OAS conversion loses fidelity;
extend scripts/patches/ as real production gaps are found.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

import yaml


def load(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    if path.suffix.lower() == ".json":
        return json.loads(text)
    return yaml.safe_load(text)


def set_servers(doc: dict[str, Any], sandbox: str, prod: str) -> None:
    doc["servers"] = [
        {"url": sandbox, "description": "Sandbox"},
        {"url": prod, "description": "Production"},
    ]


def ensure_error_schema(doc: dict[str, Any]) -> None:
    comps = doc.setdefault("components", {})
    schemas = comps.setdefault("schemas", {})
    if "FedExErrorResponse" not in schemas:
        schemas["FedExErrorResponse"] = {
            "type": "object",
            "properties": {
                "transactionId": {"type": "string"},
                "errors": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "code": {"type": "string"},
                            "message": {"type": "string"},
                        },
                    },
                },
            },
        }


def apply_json_patches(doc: dict[str, Any], patch_dir: Path) -> int:
    """Apply optional JSON Merge Patch files from scripts/patches/*.json."""
    if not patch_dir.is_dir():
        return 0
    n = 0
    for p in sorted(patch_dir.glob("*.json")):
        patch = json.loads(p.read_text(encoding="utf-8"))
        if not isinstance(patch, dict):
            continue
        # Shallow merge at top-level keys only for safety; nested handled if paths/components present
        for k, v in patch.items():
            if k == "components" and isinstance(v, dict):
                comps = doc.setdefault("components", {})
                for ck, cv in v.items():
                    if isinstance(cv, dict):
                        comps.setdefault(ck, {}).update(cv)
                    else:
                        comps[ck] = cv
            elif k == "paths" and isinstance(v, dict):
                doc.setdefault("paths", {}).update(v)
            else:
                doc[k] = v
        n += 1
        print(f"applied patch {p.name}", file=sys.stderr)
    return n


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="inp", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--server-sandbox", required=True)
    ap.add_argument("--server-prod", required=True)
    ap.add_argument("--patch-dir", default="")
    args = ap.parse_args()

    doc = load(Path(args.inp))
    if not isinstance(doc, dict):
        print("invalid OpenAPI", file=sys.stderr)
        return 1

    set_servers(doc, args.server_sandbox, args.server_prod)
    ensure_error_schema(doc)
    if args.patch_dir:
        apply_json_patches(doc, Path(args.patch_dir))

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    if out.suffix.lower() == ".json":
        out.write_text(json.dumps(doc, indent=2), encoding="utf-8")
    else:
        out.write_text(yaml.safe_dump(doc, sort_keys=False, allow_unicode=True), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
