#!/bin/bash
# StayFlow Track - Automated PostgreSQL Backup Script
# Schedule via cron: 0 */6 * * * /opt/stayflow/scripts/backup.sh
# This creates compressed backups every 6 hours and retains 7 days.

set -euo pipefail

# Configuration (override via environment)
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
DB_CONTAINER="${DB_CONTAINER:-stayflow-db}"
DB_USER="${DB_USER:-stayflow}"
DB_NAME="${DB_NAME:-stayflow_track}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
S3_BUCKET="${S3_BUCKET:-}"
S3_PREFIX="${S3_PREFIX:-stayflow-backups/postgres}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="stayflow_${TIMESTAMP}.dump"

# Ensure backup directory exists
mkdir -p "${BACKUP_DIR}"

echo "[$(date)] Starting PostgreSQL backup..."

# Create compressed custom-format dump
docker exec "${DB_CONTAINER}" pg_dump \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -Fc \
    --no-owner \
    --no-acl \
    > "${BACKUP_DIR}/${BACKUP_FILE}"

FILESIZE=$(stat -c%s "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || stat -f%z "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || wc -c < "${BACKUP_DIR}/${BACKUP_FILE}")
echo "[$(date)] Backup created: ${BACKUP_FILE} (${FILESIZE} bytes)"

# Upload to S3 if configured
if [ -n "${S3_BUCKET}" ]; then
    echo "[$(date)] Uploading to S3: s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}"
    aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}" "s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}" --quiet
    echo "[$(date)] S3 upload complete"
fi

# Remove old backups (local)
echo "[$(date)] Removing backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "stayflow_*.dump" -mtime +${RETENTION_DAYS} -delete

# Remove old S3 backups if configured
if [ -n "${S3_BUCKET}" ]; then
    CUTOFF_DATE=$(date -d "-${RETENTION_DAYS} days" +%Y-%m-%d 2>/dev/null || date -v-${RETENTION_DAYS}d +%Y-%m-%d)
    aws s3 ls "s3://${S3_BUCKET}/${S3_PREFIX}/" | while read -r line; do
        FILE_DATE=$(echo "$line" | awk '{print $1}')
        FILE_NAME=$(echo "$line" | awk '{print $4}')
        if [[ "${FILE_DATE}" < "${CUTOFF_DATE}" ]] && [[ -n "${FILE_NAME}" ]]; then
            aws s3 rm "s3://${S3_BUCKET}/${S3_PREFIX}/${FILE_NAME}" --quiet
        fi
    done
fi

echo "[$(date)] Backup complete. Retained: ${RETENTION_DAYS} days"
