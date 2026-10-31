// Package fedex is an unofficial Go client for FedEx REST APIs.
//
// # Overview
//
//   - OAuth2 client-credentials with in-memory token cache
//   - Sandbox vs Production base URLs
//   - Typed OpenAPI client via API() / AuthedClient
//
// Spec pipeline (see generate.sh):
//
//	official Postman collections → postman2openapi → merge → patch → oapi-codegen
//
// Official Postman/OpenAPI files are never committed. See DISCLAIMER.md.
package fedex
