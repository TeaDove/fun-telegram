# Telegram client utils

Telegram client utils for making telegram more fun!

## Telegram session storage

Docker Compose stores the SQLite session in the `telegram-sessions` named volume.
When running the binary directly, the default path is `.mtproto` in the working directory.
Set `FUN_TELEGRAM__SESSION_PATH` to use a different path.

### Migrate an existing Compose session

Run the following once from the directory containing the existing `.mtproto` file.
The copy stops if the volume already contains a session.

```sh
docker compose stop core
docker compose build core
docker compose run --rm --no-deps -T \
  --volume "$PWD:/legacy:ro" --entrypoint sh core -ec '
    destination=/var/lib/fun-telegram
    test -f /legacy/.mtproto
    test ! -e "$destination/.mtproto"
    for suffix in -wal -shm -journal; do
      if [ -f "/legacy/.mtproto$suffix" ]; then
        cp -p "/legacy/.mtproto$suffix" "$destination/"
      fi
    done
    cp -p /legacy/.mtproto "$destination/.mtproto.migrating"
    mv "$destination/.mtproto.migrating" "$destination/.mtproto"
  '
docker compose up -d core
```
