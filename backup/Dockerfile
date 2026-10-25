FROM postgres:17.10-alpine

# gosu ships in the postgres base for its own entrypoint; this image never uses it,
# and its old Go stdlib carried most of the image's CVEs (trivy, 2026-10-25).
RUN apk add --no-cache restic && rm -f /usr/local/bin/gosu

COPY backup.sh /usr/local/bin/neetrack-backup
COPY entrypoint.sh /usr/local/bin/neetrack-backup-entrypoint
COPY crontab /etc/crontabs/root

RUN chmod 0755 /usr/local/bin/neetrack-backup /usr/local/bin/neetrack-backup-entrypoint

ENTRYPOINT ["/usr/local/bin/neetrack-backup-entrypoint"]
