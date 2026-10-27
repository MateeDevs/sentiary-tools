# REST API reference

Explore endpoints, request parameters, responses, and models from the deployed staging API.

[Open API reference](reference.html){ .md-button .md-button--primary }

[Download the OpenAPI schema](https://docs.staging.sentiary.coolify.matee.cz/api/openapi.yaml)

The API base URL is `https://api.staging.sentiary.coolify.matee.cz/api/v1`.
The reference is read-only. Use the [backend Swagger UI](https://api.staging.sentiary.coolify.matee.cz/swagger)
or your API client to make requests.

Account and project API keys use `Authorization: Bearer <key>` on supported routes.
Account permissions and project membership still apply. Routes that require a Firebase session
use `Authorization: Bearer <Firebase ID token>`. Check each route's authentication requirements.

The schema comes from the running backend, so this reference follows deployed behavior rather
than unreleased source changes. Conditional export endpoints require the separate
[export-cache rollout](../configuration/downloads.md).
