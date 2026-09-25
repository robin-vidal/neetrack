FROM postgres:17-alpine

RUN apk add --no-cache restic

COPY backup.sh /usr/local/bin/neetrack-backup
COPY entrypoint.sh /usr/local/bin/neetrack-backup-entrypoint
COPY crontab /etc/crontabs/root

RUN chmod 0755 /usr/local/bin/neetrack-backup /usr/local/bin/neetrack-backup-entrypoint

ENTRYPOINT ["/usr/local/bin/neetrack-backup-entrypoint"]
