# REST API reference

Explore endpoints, request parameters, responses, and models from the deployed staging API.

[Open the reference in a full window](reference.html) ·
[Download the OpenAPI schema](https://docs.staging.sentiary.coolify.matee.cz/api/openapi.yaml)

The API base URL is `https://api.staging.sentiary.coolify.matee.cz/api/v1`.
The reference below is read-only. Use the [backend Swagger UI](https://api.staging.sentiary.coolify.matee.cz/swagger)
or your API client to make requests.

Account and project API keys use `Authorization: Ribbon <key>` on supported routes.
Account permissions and project membership still apply. Routes that require a Firebase session
use `Authorization: Bearer <Firebase ID token>`. Check each route's authentication requirements.

The schema comes from the running backend, so this reference follows deployed behavior rather
than unreleased source changes. Conditional export endpoints require the separate
[export-cache rollout](../configuration/downloads.md).

<iframe src="reference.html" title="Sentiary staging REST API reference" style="width: 100%; height: 80vh; min-height: 32rem; border: 0;" loading="lazy"></iframe>
