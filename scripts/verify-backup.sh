#!/bin/bash
# StayFlow Track - Backup Verification Script
# Schedule via cron: 0 3 * * 0 /opt/stayflow/scripts/verify-backup.sh
# Tests that the latest backup can be read/restored (dry-run)

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/stayflow/backups/postgres}"
HEALTHCHECK_URL="${HEALTHCHECK_PING_URL:-}"

echo "[$(date)] Starting backup verification..."

# Find latest backup
LATEST=$(find "${BACKUP_DIR}" -name "stayflow_*.dump" -type f | sort | tail -1)

if [ -z "${LATEST}" ]; then
    echo "[$(date)] ERROR: No backup files found in ${BACKUP_DIR}"
    exit 1
fi

echo "[$(date)] Verifying: ${LATEST}"

# Test backup integrity using pg_restore --list (dry-run, no DB needed)
docker exec stayflow-db pg_restore --list /dev/stdin < "${LATEST}" > /dev/null 2>&1
RESULT=$?

if [ ${RESULT} -eq 0 ]; then
    echo "[$(date)] PASS: Backup is valid and restorable"
    # Ping healthcheck if configured
    if [ -n "${HEALTHCHECK_URL}" ]; then
        curl -fsS -m 10 --retry 5 "${HEALTHCHECK_URL}" > /dev/null 2>&1 || true
    fi
else
    echo "[$(date)] FAIL: Backup verification failed (exit code: ${RESULT})"
    exit 1
fi
