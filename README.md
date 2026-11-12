# Telegram client utils

Telegram client utils for making Telegram more fun!

## Sign in with a QR code

Set the Telegram API ID and hash in `.env`. On the first interactive startup, the bot shows a QR code.
In the Telegram app on your phone, select the account you want to use, then open
**Settings → Devices → Link Desktop Device** and scan the code. The bot saves the session to `.mtproto`
in the project directory and starts normally.

For a detached Docker Compose deployment, stop the bot and run `go run . login-qr` once in an interactive
terminal from the project directory. After scanning the code, start the bot with `docker compose up -d --build`.
Subsequent starts reuse the authorized `.mtproto` session.
