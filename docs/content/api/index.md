---
tags:
  - API
---

# REST API reference

Explore endpoints, request parameters, responses, and models from the deployed API with Scalar API Reference.

[Open API reference](reference.html){ .md-button .md-button--primary }

[Download the OpenAPI schema](openapi.yaml)

The production API base URL is `https://api.sentiary.com/api/v1`.
For staging, use `https://api.staging.sentiary.coolify.matee.cz/api/v1`.
The schema download on this site follows the documentation environment. Its default
server is production; select the staging base URL explicitly when working with staging.
The reference is read-only. Use your API client to make requests.

Account and project API keys use `Authorization: Bearer <key>` on supported routes.
Account permissions and project membership still apply. Routes that require a Firebase session
use `Authorization: Bearer <Firebase ID token>`. Check each route's authentication requirements.

The reference follows the deployed API. Download the schema from the same environment as the API you use.
