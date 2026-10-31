# FedEx OpenAPI patches

Place JSON Merge Patch fragments here (applied by `scripts/preprocess.py`).

Typical fixes for Postman→OAS output:

- restore missing `required` arrays
- complete nested package / multi-piece schemas
- add enums lost in conversion
- attach `FedExErrorResponse` to error responses

Example `required-ship.json`:

```json
{
  "components": {
    "schemas": {
      "RequestedShipment": {
        "required": ["shipper", "recipients", "pickupType", "serviceType", "packagingType", "requestedPackageLineItems"]
      }
    }
  }
}
```
