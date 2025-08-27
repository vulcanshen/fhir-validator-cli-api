# FHIR Validator CLI API

This project provides a RESTful API interface for the [FHIR validator_cli.jar](https://confluence.hl7.org/display/FHIR/Using+the+FHIR+Validator) using Go and the Gin web framework. It allows users to validate FHIR resources by sending HTTP requests, making it easier to integrate FHIR validation into automated workflows and other systems.

## Features
- Exposes validator_cli.jar as a web API
- Synchronous and Server-Sent Events (SSE) validation endpoints
- Swagger/OpenAPI documentation
- Simple deployment via Docker

## Endpoints

### Synchronous Validation
- `POST /api/validate`
  - Validates a FHIR payload and returns the result as plain text.

### SSE Validation
- `POST /api/validate/sse`
  - Streams the validation output using Server-Sent Events.

## Request Example
```json
{
  "payload": { ...FHIR resource... },
  "args": ["-version", "4.0.1"]
}
```

## Usage
1. Place `validator_cli.jar` in the project root.
2. Build and run the API:
   ```sh
   go build -o fhir-validator-api
   ./fhir-validator-api
   ```
3. Access Swagger UI at `http://localhost:8081/swagger/index.html`

## Docker
To run with Docker:

```sh
docker run -p 8081:8081 vulcanshen2304/fhir-validator-cli-api:latest
```

## Development
- Requires Go 1.18+
- Uses [Gin](https://github.com/gin-gonic/gin) and [swaggo](https://github.com/swaggo/swag) for API and documentation

## License
See [LICENSE](LICENSE).

## Author
- Vulcan Shen ([GitHub](https://github.com/vulcanshen))
