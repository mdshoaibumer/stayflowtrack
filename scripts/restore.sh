#!/bin/bash
# StayFlow Track - PostgreSQL Restore Script
# Usage: ./restore.sh /backups/postgres/stayflow_20260605_060000.dump

set -euo pipefail

if [ -z "${1:-}" ]; then
    echo "Usage: $0 <backup_file>"
    echo "Example: $0 /backups/postgres/stayflow_20260605_060000.dump"
    exit 1
fi

BACKUP_FILE="$1"
DB_CONTAINER="${DB_CONTAINER:-stayflow-db}"
DB_USER="${DB_USER:-stayflow}"
DB_NAME="${DB_NAME:-stayflow_track}"

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "ERROR: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

echo "[$(date)] Starting restore from: ${BACKUP_FILE}"
echo "WARNING: This will drop and recreate the database '${DB_NAME}'."
read -p "Are you sure? (yes/no): " CONFIRM

if [ "${CONFIRM}" != "yes" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Drop existing connections and recreate database
docker exec "${DB_CONTAINER}" psql -U "${DB_USER}" -d postgres -c "
    SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();
"
docker exec "${DB_CONTAINER}" psql -U "${DB_USER}" -d postgres -c "DROP DATABASE IF EXISTS ${DB_NAME};"
docker exec "${DB_CONTAINER}" psql -U "${DB_USER}" -d postgres -c "CREATE DATABASE ${DB_NAME} OWNER ${DB_USER};"

# Restore from dump
cat "${BACKUP_FILE}" | docker exec -i "${DB_CONTAINER}" pg_restore \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --no-owner \
    --no-acl \
    --clean \
    --if-exists

echo "[$(date)] Restore complete."
