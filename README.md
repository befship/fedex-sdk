# fedex-sdk

English | [中文](./README.zh-CN.md)

Unofficial Go SDK for [FedEx REST APIs](https://developer.fedex.com/), produced via **Postman Collection → OpenAPI 3 → oapi-codegen**, with OAuth, retries, and a stable facade.

> **Not affiliated with FedEx.** You need your own FedEx Developer Portal account and OAuth credentials. See [DISCLAIMER.md](./DISCLAIMER.md).

## Why this layout

| Path | Purpose |
|------|---------|
| `fedex/` | Hand-written facade (OAuth cache, env URLs, typed API access) |
| `internal/generated/fedexapi/` | `oapi-codegen` output (committed so `go get` works) |
| `scripts/` + `generate.sh` | Postman → OAS → merge → patch → generate |
| `.cache/` | Downloaded Postman / OpenAPI (**gitignored**, never commit) |

FedEx does **not** publish a single complete OpenAPI YAML. The automation path is:

```
official Postman (Rate / Ship / Track)
  → postman2openapi
  → merge_oas.py
  → preprocess.py (+ scripts/patches)
  → oapi-codegen
```

## Install

```bash
go get github.com/befship/fedex-sdk@latest
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/fedex-sdk/fedex"
)

func main() {
	client := fedex.New(fedex.Config{
		ClientID:     os.Getenv("FEDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("FEDEX_CLIENT_SECRET"),
		Environment:  fedex.EnvSandbox,
	})
	tok, err := client.AccessToken(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("token ok, len=", len(tok))
}
```

## Endpoints

| Environment | Base URL |
|---|---|
| Sandbox | `https://apis-sandbox.fedex.com` |
| Production | `https://apis.fedex.com` |

## Regenerate

1. From [FedEx API catalog](https://developer.fedex.com/api/en-us/catalog.html) download Postman collections for **Rate**, **Ship**, and **Track**.
2. Rename to `fedex-rate.json`, `fedex-ship.json`, `fedex-track.json`.
3. Run:

```bash
npm install -g postman2openapi
pip3 install pyyaml
POSTMAN_DIR=/path/to/collections make generate
```

Or bootstrap from an already-merged OpenAPI:

```bash
SPEC_LOCAL=/path/to/openapi.json make generate
```

## License

Hand-written code: MIT. Generated clients: subject to FedEx API terms. See `LICENSE` and `DISCLAIMER.md`.
