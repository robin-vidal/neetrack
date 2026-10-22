#!/bin/sh
set -eu

: "${RESTIC_REPOSITORY:?Set RESTIC_REPOSITORY in backup.env}"
: "${RESTIC_PASSWORD:?Set RESTIC_PASSWORD in backup.env}"
: "${AWS_ACCESS_KEY_ID:?Set AWS_ACCESS_KEY_ID in backup.env}"
: "${AWS_SECRET_ACCESS_KEY:?Set AWS_SECRET_ACCESS_KEY in backup.env}"

mkdir -p "${BACKUP_DIR:-/backups}"
lock_dir=/tmp/neetrack-backup.lock
if ! mkdir "$lock_dir" 2>/dev/null; then
  exit 0
fi
trap 'rmdir "$lock_dir"' EXIT

stamp=$(date -u +%Y%m%dT%H%M%SZ)
dump_file="${BACKUP_DIR:-/backups}/neetrack-${stamp}.dump"
pg_dump --format=custom --file="$dump_file"

if ! restic snapshots >/dev/null 2>&1; then
  restic init
fi

restic backup "$dump_file" --host neetrack --tag postgresql
restic forget --host neetrack --keep-daily 7 --prune
find "${BACKUP_DIR:-/backups}" -type f -name 'neetrack-*.dump' -mtime +7 -delete
