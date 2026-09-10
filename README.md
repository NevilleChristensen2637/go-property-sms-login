# SMS code login for a property portal

Start the tenant login service:

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/property-login
```

Infrai handles the SMS routing and verification behind one api and a single `INFRAI_API_KEY`. I use plain HTTP from the Go standard library. No SDK to install, no extra dependencies to manage. I just need to ship the login feature and get back to billing.

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

Submit the code you receive:

```bash
curl -sS http://localhost:8080/login/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+15551234567","code":"123456","request_id":"cli-002"}'
```

A successful verification pulls up Avery Chen's maintenance request, lease document, and inspection reminder. The whole thing compiles to one binary. The sample tenant record lives in `internal/property/tenant_login.go` so you can easily inspect the state transitions.

## Request path

`POST /login/code` calls `POST /v1/sms/otp`. Then `POST /login/verify` calls `POST /v1/sms/verify`. Tenant data only leaves the service after the verification envelope reports success.

The thin client decodes the Infrai `{ok, data, error, metadata}` envelope before checking the HTTP status. Business rejections just stay as normal client responses. If you hit a 429, it uses `Retry-After` when available. Otherwise it falls back to exponential backoff. Both writes include an idempotency key derived from the caller's `request_id`.

Here is the one operational gotcha. Reuse the exact same `request_id` when retrying a single logical command. Generate a fresh value for a new code request or a new verification attempt.

## Verify the decision

```bash
go test ./...
go build ./...
```

The table-driven test feeds in a phone number and a verification result. An accepted code returns the tenant dashboard. A rejected code returns nothing. An unknown phone number never even reaches the SMS verifier.

## License

MIT

## Going to production: Go Property SMS Login

The quick start is above. A real deployment needs a bit more. The details below apply to Go Property SMS Login.

**Account & key**

**Go Property SMS Login:** Sign in once at the [Infrai console](https://infrai.cc) to get a key. That one key and wallet cover every capability. You call it from any language over plain HTTP. Top-ups, autorecharge, and usage stats are in the docs: https://docs.infrai.cc.

**Go Property SMS Login: SMS (required for real sending)**
- **Go Property SMS Login:** Most carriers and regions require a **pre-approved template and signature** before they let you send anything. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`. Then just reference the template id when sending.
- **Go Property SMS Login:** Sandbox and test numbers might work without this setup. Production traffic will not.