---
tags:
  - API
---

# REST API reference

Explore endpoints, request parameters, responses, and models from the deployed staging API with Scalar API Reference.

[Open API reference](reference.html){ .md-button .md-button--primary }

[Download the OpenAPI schema](https://docs.staging.sentiary.coolify.matee.cz/api/openapi.yaml)

The API base URL is `https://api.staging.sentiary.coolify.matee.cz/api/v1`.
The reference is read-only. Use the [backend Swagger UI](https://api.staging.sentiary.coolify.matee.cz/swagger)
or your API client to make requests.

Account and project API keys use `Authorization: Bearer <key>` on supported routes.
Account permissions and project membership still apply. Routes that require a Firebase session
use `Authorization: Bearer <Firebase ID token>`. Check each route's authentication requirements.

The reference follows the deployed API. Download the schema from the same environment as the API you use.
