#!/bin/bash
# StayFlow Track - MinIO (Object Storage) Backup Script
# Schedule via cron: 0 2 * * * /opt/stayflow/scripts/backup-minio.sh
# Backs up uploaded documents (guest IDs, invoices, etc.)

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/stayflow/backups/minio}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="minio_${TIMESTAMP}.tar.gz"

mkdir -p "${BACKUP_DIR}"

echo "[$(date)] Starting MinIO backup..."

# Export MinIO volume data as compressed archive
docker run --rm \
  -v stayflow-track_minio_data:/data:ro \
  -v "${BACKUP_DIR}:/backup" \
  alpine tar czf "/backup/${BACKUP_FILE}" -C /data .

FILESIZE=$(stat -c%s "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || stat -f%z "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || wc -c < "${BACKUP_DIR}/${BACKUP_FILE}")
echo "[$(date)] MinIO backup created: ${BACKUP_FILE} (${FILESIZE} bytes)"

# Remove old backups
find "${BACKUP_DIR}" -name "minio_*.tar.gz" -mtime +${RETENTION_DAYS} -delete

echo "[$(date)] MinIO backup complete. Retained: ${RETENTION_DAYS} days"
