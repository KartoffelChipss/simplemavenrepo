#!/bin/sh
set -e

if [ "$(id -u)" = "0" ]; then
    PUID="${PUID:-1000}"
    PGID="${PGID:-1000}"

    for dir in "$SMR_REPO_PATH" "$SMR_CONFIG_PATH"; do
        [ -n "$dir" ] || continue
        mkdir -p "$dir"
        find "$dir" \( ! -user "$PUID" -o ! -group "$PGID" \) -exec chown -h "$PUID:$PGID" {} +
    done

    exec su-exec "$PUID:$PGID" simplemavenrepo "$@"
fi

exec simplemavenrepo "$@"
