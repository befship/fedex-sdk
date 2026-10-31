#!/usr/bin/env python3
"""Merge multiple OpenAPI 3 YAML/JSON docs into one (paths + components)."""
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
        doc = json.loads(text)
    else:
        doc = yaml.safe_load(text)
    if not isinstance(doc, dict):
        raise SystemExit(f"invalid OpenAPI: {path}")
    return doc


def merge_dicts(dst: dict[str, Any], src: dict[str, Any], label: str) -> None:
    for k, v in src.items():
        if k not in dst:
            dst[k] = v
            continue
        if dst[k] == v:
            continue
        # Prefer first definition; warn on conflict
        print(f"warn: duplicate {label}.{k} — keeping first", file=sys.stderr)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", required=True)
    ap.add_argument("--title", default="FedEx REST APIs (merged)")
    ap.add_argument("--server-sandbox", required=True)
    ap.add_argument("--server-prod", required=True)
    ap.add_argument("inputs", nargs="+", help="OpenAPI YAML/JSON files")
    args = ap.parse_args()

    docs = [load(Path(p)) for p in args.inputs]
    base = {
        "openapi": docs[0].get("openapi", "3.0.3"),
        "info": {
            "title": args.title,
            "version": str((docs[0].get("info") or {}).get("version") or "1.0.0"),
            "description": "Merged from official FedEx Postman collections (unofficial).",
        },
        "servers": [
            {"url": args.server_sandbox, "description": "Sandbox"},
            {"url": args.server_prod, "description": "Production"},
        ],
        "paths": {},
        "components": {"schemas": {}, "securitySchemes": {}, "parameters": {}, "responses": {}, "requestBodies": {}},
    }

    for doc in docs:
        merge_dicts(base["paths"], doc.get("paths") or {}, "paths")
        comps = doc.get("components") or {}
        for section in ("schemas", "securitySchemes", "parameters", "responses", "requestBodies"):
            merge_dicts(base["components"][section], comps.get(section) or {}, f"components.{section}")
        if doc.get("security") and not base.get("security"):
            base["security"] = doc["security"]

    # Drop empty component sections
    base["components"] = {k: v for k, v in base["components"].items() if v}

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    if out.suffix.lower() == ".json":
        out.write_text(json.dumps(base, indent=2), encoding="utf-8")
    else:
        out.write_text(yaml.safe_dump(base, sort_keys=False, allow_unicode=True), encoding="utf-8")
    print(f"→ merged {len(docs)} specs → {out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
