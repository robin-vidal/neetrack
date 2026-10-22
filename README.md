# Neetrack

Neetrack is a self-hosted Neetcode 150 tracker built with Go, React, and PostgreSQL.

## Backups

The optional `backup` Compose profile creates a compressed PostgreSQL dump every day at
03:17 UTC, stores seven days of dumps under `./backups`, and encrypts snapshots with Restic
before uploading them to Backblaze B2. The backup container also runs a backup when it starts.

Create a private B2 bucket and a bucket-scoped application key with read, write, and delete
access. Set the bucket lifecycle rule to keep only the latest version of each file so replaced
Restic pack files do not keep consuming storage. B2's S3-compatible endpoint and region are
shown in the bucket details. Copy the example config, replace its placeholders, and protect
the credentials file:

```sh
cp backup.env.example backup.env
chmod 600 backup.env
```

Set `RESTIC_REPOSITORY` to
`s3:https://s3.<region>.backblazeb2.com/<bucket>/neetrack`, use the B2 key ID as
`AWS_ACCESS_KEY_ID`, the application key as `AWS_SECRET_ACCESS_KEY`, set `AWS_DEFAULT_REGION`
to the bucket region, and set a unique, strong `RESTIC_PASSWORD`. Keep that password separately
from the bucket key because Restic uses it to encrypt the backup repository.

Start the backup service from the directory containing `compose.yaml`:

```sh
docker compose --profile backup up -d --build backup
docker compose logs -f backup
```

Restic retains seven daily snapshots and prunes data no longer referenced by those snapshots.
The local dump directory also removes files older than seven days. Keep the B2 bucket private.

### Restore check

Run this after configuring backups and at least once before relying on them. It restores the
latest snapshot into a temporary PostgreSQL database that uses a memory-backed data directory,
then starts Neetrack against that restored database:

```sh
docker compose --profile backup --profile restore-check up -d --wait restore-db backup
docker compose --profile backup exec backup restic snapshots
docker compose --profile backup exec backup sh -c 'restic restore latest --target /tmp/neetrack-restore'
docker compose --profile backup exec -e PGHOST=restore-db -e PGPASSWORD=restore-check backup sh -c 'dump=$(find /tmp/neetrack-restore -type f -name "neetrack-*.dump" | sort | tail -n 1); pg_restore --create --clean --if-exists --dbname=postgres "$dump"'
docker compose run --rm --no-deps -p 127.0.0.1:18080:8080 -e DATABASE_URL='postgres://neetrack:restore-check@restore-db:5432/neetrack?sslmode=disable' app
```

In another terminal, verify `curl -fsS http://127.0.0.1:18080/api/health` returns
`{"status":"ok"}`. Stop the foreground app with Ctrl+C, then stop the temporary database:

```sh
docker compose --profile restore-check stop restore-db
```

The temporary database disappears when it stops. This procedure does not write to the live
database.
