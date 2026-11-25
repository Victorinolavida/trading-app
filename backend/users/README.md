# Local password reset with Mailpit

From the project root, start Mailpit with `docker compose up -d mailpit` and open
http://localhost:8025. Copy `config/config.example.yaml` to `config/config.yaml`
if needed, then set `env: development`, `password_reset.enabled: true`, and
`password_reset.provider: mailpit`. The example's Mailpit host, port and sender
work when the users service runs on the host. If it runs in Docker on the same
network as Mailpit, use `mailpit` as the host instead.

Start PostgreSQL, run `just migrate`, and start the service with `just run`.
Request a reset for an existing user (requires `grpcurl`):

```sh
grpcurl -plaintext -d '{"email":"user@example.com"}' localhost:3030 auth.AuthService/RequestPasswordReset
```

Open the email in Mailpit and copy the `token` query parameter from the reset
link. The configured reset URL must point to a running frontend if you want to
open the link directly. To confirm through gRPC instead:

```sh
grpcurl -plaintext -d '{"token":"PASTE_TOKEN_HERE","newPassword":"NewPassword123"}' localhost:3030 auth.AuthService/ConfirmPasswordReset
```

The token is single-use and expires after the configured TTL. Reset requests
return the same response for unknown users and delivery errors; check the users
service logs if an expected email does not appear. Stop Mailpit with
`docker compose stop mailpit` from the project root.

## Production email

Set `env: production`, choose `password_reset.provider: smtp`, set the reset
URL to your public HTTPS frontend, and configure `password_reset.smtp` with your
provider's SMTP host, STARTTLS port (usually 587), and verified `from` address.
Provide credentials through the `RESET_SMTP_USERNAME` and `RESET_SMTP_PASSWORD`
environment variables using your deployment's secret manager. The SMTP sender
requires STARTTLS with certificate verification and SMTP PLAIN authentication;
it does not support implicit TLS on port 465 or provider-specific APIs. Check
that your chosen provider supports these settings. Configure your sender domain
with the provider's required DNS records before sending real mail.

The `mock` and `mailpit` providers remain development-only. Restart the service
after changing its configuration.
