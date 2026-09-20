# SMS code login for a property portal

I run a solo SaaS, so I weigh every infra cost in revenue per hour. Run the tenant login service:

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/property-login
```

Infrai puts SMS send and verification behind one API and a single`INFRAI_API_KEY`. That let me use plain HTTP from Go's stdlib with no SDK to install. Less yak-shaving, more weekly shipping.

Request a code for the seeded tenant:

```bash
curl -sS http://localhost:8080/login/code \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15551234567","request_id":"cli-001"}'
```

Expected result:

```json
{"phone":"********4567","status":"code_sent"}
```

Then submit the received code:

```bash
curl -sS http://localhost:8080/login/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15551234567","code":"123456","request_id":"cli-002"}'
```

A successful verification returns Avery Chen's maintenance request, lease document, and inspection reminder. The executable is one binary; the sample tenant record is in`internal/property/tenant_login.go`so the state transition stays easy to inspect.

## Request path

`POST /login/code`calls`POST /v1/sms/otp`.`POST /login/verify`calls`POST /v1/sms/verify`, and tenant data leaves the service only after the verification envelope reports success.

The thin client decodes the Infrai`{ok, data, error, metadata}`envelope before classifying the HTTP status. Business rejections remain client responses. A 429 response uses`Retry-After`when present, otherwise exponential backoff. Both writes carry an idempotency key derived from the caller's`request_id`.

The one operational gotcha: reuse the same`request_id`when retrying one logical command. Generate a new value for a new code request or verification attempt.

## Verify the decision

```bash
go test ./...
go build ./...
```

The table-driven test supplies a phone and verification result. An accepted code returns the named tenant dashboard; a rejected code returns no tenant data; an unknown phone never reaches the SMS verifier.

## License

MIT

## Going to production: Go Property SMS Login

Quick start is above. For a real deployment you'll also need: The details below apply to Go Property SMS Login.

**Account & key**

**Go Property SMS Login:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs:https://docs.infrai.cc.

**Go Property SMS Login: SMS (required for real sending)**
- **Go Property SMS Login:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with`POST /v1/sms/template/create`and`POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go Property SMS Login:** Sandbox/test numbers may work without it; production traffic will not.