# fedex-sdk

[English](./README.md) | 中文

非官方 Go SDK。FedEx **不提供**统一完整 OpenAPI，因此采用：

**官方 Postman Collection → OAS3 → 合并 → Patch → oapi-codegen**

并封装 OAuth、重试与稳定对外接口。

> **与 FedEx 无隶属关系。** 使用前请自行注册 FedEx Developer Portal。详见 [DISCLAIMER.md](./DISCLAIMER.md)。

## 重新生成

```bash
# 从开发者门户下载 Rate / Ship / Track 三个 Postman Collection
POSTMAN_DIR=/path/to/collections make generate

# 或直接给已合并的 OpenAPI
SPEC_LOCAL=/path/to/openapi.json make generate
```

官方 Postman / OpenAPI **不会**提交进 Git。
