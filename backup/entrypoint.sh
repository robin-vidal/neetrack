#!/bin/sh
set -eu

/usr/local/bin/neetrack-backup
exec crond -f -l 2
