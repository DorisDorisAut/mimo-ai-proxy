# Wina DEV compatibility

MiMo Studio moved generation from `bot/chat` to `react/chat`. An error event with HTTP 200 is not a valid empty completion.

Instance settings: `MIMO_CHAT_PROTOCOL=react`, `MIMO_STATELESS=1`, `LISTEN_HOST=127.0.0.1`.

Stateless mode creates a fresh provider conversation per request, forwards supplied context, and disables history fingerprint recovery and response caching. Keep company authorization in Django.

Run `go test ./...`; coverage includes content/reasoning events, empty/error responses, streaming error masking and separation of company contexts. Back up binary/environment, merge settings, validate a separate loopback instance with a disposable history DB, then validate normal/streaming responses and the Django helper on the active proxy. Do not restart financial services for this proxy-only change.

Existing Studio credentials stay server-side. No new paid API provider is required. Session validity and quota remain external dependencies.
