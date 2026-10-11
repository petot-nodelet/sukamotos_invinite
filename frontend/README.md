# SUKAMOTO frontend

Static HTML, CSS, and JavaScript frontend. No framework or npm dependencies are required.

## Local development

Run the Go backend from `backend/` in one terminal:

```sh
go run ./cmd/server
```

Run the static frontend server from the repository root in another terminal:

```sh
node frontend/dev-server.mjs
```

Open `http://localhost:3000`. On that exact local origin, `js/config.js` points the API client to `http://localhost:8080/api/v1`. For other development origins, edit `apiBaseUrl` in `js/config.js` and configure the backend's `ALLOWED_ORIGIN` to the frontend origin. Do not put credentials or secrets in the frontend configuration.

For same-origin production hosting, `apiBaseUrl` defaults to `/api/v1`. Serve the frontend and API under the same origin where possible. The browser API client sends session cookies with `credentials: "include"` and reads the configured CSRF cookie name (`csrfCookieName`) to send `X-CSRF-Token` on mutating requests.

## Pages and API usage

- `index.html`: public package and template catalog; creates orders and opens private, data-filled previews in an iframe.
- `status.html`: looks up a public order reference and submits private payment proof.
- `admin/login.html` and `admin/index.html`: cookie-based admin login, orders, payment decisions, invitation publication, package read-only view, and template availability.
- `js/api.js`: shared JSON/multipart client, timeout, response-envelope validation, error normalization, and endpoint methods.
- `templates/`: all six original invitation designs plus `personalize.js`, the shared text-safe bridge used for previews and published invitations. The invitation CSS, effects, audio, galleries, flipbook and original scripts remain in their individual HTML files.

The form loads `supported_fields` from `GET /api/v1/templates/{slug}` before accepting an order. The data-filled preview loads the selected demo in an iframe and sends form values with origin-checked `postMessage`; it does not place private values in the URL or persistent browser storage. Published data is embedded as JSON by the backend and applied using `textContent`. Prices and final order details are always displayed from API responses. Admin buttons are gated by order state for usability only; authorization remains a backend responsibility.

## Integration constraints found in the current backend

- The API contract has no package mutation endpoint, so package management is read-only.
- The admin detail response has payment metadata but no authorized proof retrieval route; the UI never constructs or exposes a proof URL.
- Database seeding registers all six audited files in `frontend/templates/` and aligns their source/demo metadata. The existing three catalog slugs are preserved to avoid breaking old orders.
- The backend serves template previews at `/demo/...`, the shared runtime at `/templates/personalize.js`, and published original HTML at `/i/{slug}`. Restart the backend after updating code so seed metadata and static routes are refreshed.
- The API contract requires CSRF protection for cookie-authenticated admin mutations, but the current backend README identifies that protection as pending. Frontend headers do not replace server-side CSRF validation.

## Validation

No browser automation or mobile-device emulation is bundled. Use browser devtools on narrow/mobile and desktop viewports and verify the Go API's CORS configuration for any split-origin setup.
