# StayFlow Track — Production Deployment Guide

> Optimized for: Single VPS, Low Cost, Solo Founder, 99.9% Uptime  
> Last reviewed: 2026-06-05

---

## Table of Contents

1. [Production Deployment Checklist](#1-production-deployment-checklist)
2. [Environment Variables](#2-environment-variables)
3. [Backup Strategy](#3-backup-strategy)
4. [Restore Strategy](#4-restore-strategy)
5. [Monitoring Setup](#5-monitoring-setup)
6. [Security Checklist](#6-security-checklist)
7. [Findings & Recommendations](#7-findings--recommendations)

---

## 1. Production Deployment Checklist

### Pre-Deployment (One-time VPS Setup)

| # | Task | Status |
|---|------|--------|
| 1 | Provision VPS (minimum 4GB RAM, 2 vCPU, 80GB SSD) | ☐ |
| 2 | Ubuntu 24.04 LTS installed | ☐ |
| 3 | Create non-root user with sudo | ☐ |
| 4 | Disable root SSH login, set `PermitRootLogin no` | ☐ |
| 5 | Configure SSH key-only auth, disable password auth | ☐ |
| 6 | Install Docker Engine + Docker Compose plugin | ☐ |
| 7 | Configure UFW firewall: allow 22, 80, 443 only | ☐ |
| 8 | Enable automatic security updates (`unattended-upgrades`) | ☐ |
| 9 | Set timezone to UTC (`timedatectl set-timezone UTC`) | ☐ |
| 10 | Configure swap (2GB) for memory safety | ☐ |
| 11 | Set up DNS A records for `api.yourdomain.com` + `app.yourdomain.com` | ☐ |
| 12 | Install `fail2ban` for SSH brute-force protection | ☐ |

### Application Deployment

| # | Task | Status |
|---|------|--------|
| 13 | Clone repo to `/opt/stayflow/` | ☐ |
| 14 | Create `.env` file from `.env.example` with production values | ☐ |
| 15 | Generate strong JWT secrets (`openssl rand -base64 48`) | ☐ |
| 16 | Generate strong DB password (`openssl rand -base64 32`) | ☐ |
| 17 | Generate MinIO credentials (`openssl rand -base64 24`) | ☐ |
| 18 | Build images: `docker compose -f docker-compose.prod.yml build` | ☐ |
| 19 | Run migrations: `docker compose -f docker-compose.prod.yml run --rm migrate` | ☐ |
| 20 | Start stack: `docker compose -f docker-compose.prod.yml up -d` | ☐ |
| 21 | Verify SSL: `curl -I https://api.yourdomain.com/health` | ☐ |
| 22 | Verify frontend: `curl -I https://app.yourdomain.com` | ☐ |
| 23 | Test readiness: `curl https://api.yourdomain.com/ready` → `{"ready":true}` | ☐ |

### Post-Deployment

| # | Task | Status |
|---|------|--------|
| 24 | Set up cron backup schedule (see §3) | ☐ |
| 25 | Configure log rotation (see §5) | ☐ |
| 26 | Set up uptime monitoring (UptimeRobot / Healthchecks.io) | ☐ |
| 27 | Test backup + restore on staging data | ☐ |
| 28 | Set up systemd service for Docker Compose auto-start | ☐ |
| 29 | Document runbooks for common failure scenarios | ☐ |
| 30 | Register first admin tenant account | ☐ |

---

## 2. Environment Variables

### Required `.env` file for production

```bash
# ─── Application ───────────────────────────────────────────────
APP_ENV=production
APP_PORT=8080

# ─── Domain (Caddy auto-SSL) ──────────────────────────────────
DOMAIN_API=api.yourdomain.com
DOMAIN_APP=app.yourdomain.com

# ─── PostgreSQL ────────────────────────────────────────────────
DB_USER=stayflow
DB_PASSWORD=<openssl rand -base64 32>
DB_NAME=stayflow_track
DB_HOST=postgres
DB_PORT=5432
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_MAX_LIFETIME=5m
DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable

# ─── JWT Authentication ───────────────────────────────────────
JWT_ACCESS_SECRET=<openssl rand -base64 48>
JWT_REFRESH_SECRET=<openssl rand -base64 48>
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h
JWT_ISSUER=stayflow-track

# ─── Object Storage (MinIO) ───────────────────────────────────
STORAGE_PROVIDER=s3
STORAGE_ENDPOINT=http://minio:9000
STORAGE_REGION=us-east-1
STORAGE_BUCKET=stayflow-documents
STORAGE_ACCESS_KEY_ID=<openssl rand -base64 24>
STORAGE_SECRET_ACCESS_KEY=<openssl rand -base64 32>
STORAGE_USE_PATH_STYLE=true

# ─── CORS ─────────────────────────────────────────────────────
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com

# ─── Logging ──────────────────────────────────────────────────
LOG_LEVEL=info
LOG_FORMAT=json

# ─── Razorpay (payments — optional) ───────────────────────────
RAZORPAY_KEY_ID=
RAZORPAY_KEY_SECRET=
RAZORPAY_WEBHOOK_SECRET=

# ─── SaaS Platform ────────────────────────────────────────────
PLATFORM_TENANT_ID=<uuid-of-platform-admin-tenant>

# ─── Backup (set in cron environment) ─────────────────────────
BACKUP_DIR=/opt/stayflow/backups
RETENTION_DAYS=7
S3_BUCKET=
S3_PREFIX=stayflow-backups/postgres
```

### Secret Generation Commands

```bash
# JWT secrets (minimum 32 chars enforced in production)
openssl rand -base64 48  # access secret
openssl rand -base64 48  # refresh secret (must differ)

# Database password
openssl rand -base64 32

# MinIO credentials
openssl rand -base64 24  # access key
openssl rand -base64 32  # secret key
```

---

## 3. Backup Strategy

### Current Implementation: `scripts/backup.sh`

| Aspect | Configuration |
|--------|--------------|
| Format | PostgreSQL custom format (`pg_dump -Fc`) |
| Schedule | Every 6 hours via cron |
| Retention | 7 days local |
| Offsite | Optional S3/compatible storage |
| Data covered | PostgreSQL only |

### Recommended Cron Schedule

```bash
# Edit crontab: crontab -e
# PostgreSQL backup every 6 hours
0 */6 * * * /opt/stayflow/scripts/backup.sh >> /var/log/stayflow-backup.log 2>&1

# MinIO data backup daily at 2 AM
0 2 * * * /opt/stayflow/scripts/backup-minio.sh >> /var/log/stayflow-minio-backup.log 2>&1

# Backup verification weekly (Sunday at 3 AM)
0 3 * * 0 /opt/stayflow/scripts/verify-backup.sh >> /var/log/stayflow-backup-verify.log 2>&1
```

### Missing: MinIO Backup Script

Create `scripts/backup-minio.sh`:

```bash
#!/bin/bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/opt/stayflow/backups/minio}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "${BACKUP_DIR}"

# Use docker cp to export MinIO volume
docker run --rm \
  -v stayflow-track_minio_data:/data:ro \
  -v "${BACKUP_DIR}:/backup" \
  alpine tar czf "/backup/minio_${TIMESTAMP}.tar.gz" -C /data .

# Retention
find "${BACKUP_DIR}" -name "minio_*.tar.gz" -mtime +7 -delete
echo "[$(date)] MinIO backup complete"
```

### 3-2-1 Backup Rule (Low-Cost)

| Copy | Location | Cost |
|------|----------|------|
| 1 | VPS local (`/opt/stayflow/backups/`) | Free |
| 2 | S3-compatible (Backblaze B2, Wasabi, or Hetzner Storage Box) | ~$1/month |
| 3 | Download to local machine weekly (optional) | Free |

### WAL Archiving (Point-in-Time Recovery)

For 99.9% uptime with minimal data loss, add PostgreSQL WAL archiving:

```yaml
# Add to postgres service in docker-compose.prod.yml
command: >
  postgres
  -c wal_level=replica
  -c archive_mode=on
  -c archive_command='cp %p /backups/wal/%f'
  -c max_wal_senders=2
volumes:
  - postgres_data:/var/lib/postgresql/data
  - wal_backups:/backups/wal
```

This enables recovery to any point in time between full backups (RPO ≈ seconds instead of 6 hours).

---

## 4. Restore Strategy

### Restore Runbook

#### Scenario A: Application Crash (most common)

```bash
cd /opt/stayflow
docker compose -f docker-compose.prod.yml restart backend
# Verify:
curl https://api.yourdomain.com/health
```

Recovery time: **< 30 seconds**

#### Scenario B: Database Corruption / Data Loss

```bash
# 1. Stop application
docker compose -f docker-compose.prod.yml stop backend frontend

# 2. Find latest backup
ls -la /opt/stayflow/backups/postgres/

# 3. Restore (interactive confirmation required)
./scripts/restore.sh /opt/stayflow/backups/postgres/stayflow_YYYYMMDD_HHMMSS.dump

# 4. Re-run any newer migrations if needed
docker compose -f docker-compose.prod.yml run --rm migrate

# 5. Restart application
docker compose -f docker-compose.prod.yml up -d backend frontend

# 6. Verify
curl https://api.yourdomain.com/ready
```

Recovery time: **5-15 minutes** (depending on DB size)

#### Scenario C: Full VPS Failure (Disaster Recovery)

```bash
# On new VPS:
# 1. Run pre-deployment checklist (§1)
# 2. Clone repo
git clone <repo-url> /opt/stayflow && cd /opt/stayflow

# 3. Restore .env (keep a copy in password manager)
cp /secure-location/.env .env

# 4. Download latest backup from S3
aws s3 cp s3://${S3_BUCKET}/${S3_PREFIX}/latest.dump /opt/stayflow/backups/

# 5. Start DB only
docker compose -f docker-compose.prod.yml up -d postgres
# Wait for healthy
docker compose -f docker-compose.prod.yml exec postgres pg_isready

# 6. Restore database
./scripts/restore.sh /opt/stayflow/backups/latest.dump

# 7. Start everything
docker compose -f docker-compose.prod.yml up -d

# 8. Update DNS to new VPS IP
# 9. Wait for SSL provisioning (Caddy auto-provisions)
```

Recovery time: **30-60 minutes**

#### Scenario D: Rollback Bad Deployment

```bash
# 1. Identify last good image
docker images | grep stayflow

# 2. Rollback to previous version
git checkout <previous-tag>
docker compose -f docker-compose.prod.yml up -d --build backend

# 3. If migration needs rollback
docker compose -f docker-compose.prod.yml run --rm migrate \
  sh -c "migrate -path /app/migrations -database '${DATABASE_URL}' down 1"
```

### Recovery Time Objectives (RTO/RPO)

| Scenario | RTO | RPO |
|----------|-----|-----|
| App crash | 30s | 0 (no data loss) |
| DB restore | 15 min | 6 hours (last backup) |
| Full VPS rebuild | 60 min | 6 hours |
| With WAL archiving | 60 min | ~5 seconds |

---

## 5. Monitoring Setup

### Tier 1: Free External Uptime Monitoring (Mandatory)

**UptimeRobot** (free tier: 50 monitors, 5-min checks)

| Monitor | URL | Check Interval |
|---------|-----|----------------|
| API Health | `https://api.yourdomain.com/health` | 5 min |
| API Ready | `https://api.yourdomain.com/ready` | 5 min |
| Frontend | `https://app.yourdomain.com` | 5 min |

Configure alerts: Email + Telegram/Slack webhook.

### Tier 2: Healthchecks.io for Cron Monitoring (Free)

Register backup cron jobs with [Healthchecks.io](https://healthchecks.io):

```bash
# Add to end of backup.sh
curl -fsS -m 10 --retry 5 "https://hc-ping.com/<your-uuid>" > /dev/null
```

If backup doesn't ping within expected window → alert fires.

### Tier 3: On-VPS Log Monitoring

#### Log Rotation (create `/etc/logrotate.d/stayflow`)

```
/var/log/stayflow-*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 root root
}
```

#### Docker Log Rotation (create `/etc/docker/daemon.json`)

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "5"
  }
}
```

Then restart Docker: `systemctl restart docker`

### Tier 4: Lightweight Metrics (Optional, Low-Cost)

Your app already exposes basic metrics via `/health` (active_requests, total_requests, total_errors). Add a simple collector:

#### Option A: `scripts/collect-metrics.sh` (Zero-cost)

```bash
#!/bin/bash
# Cron: */5 * * * * /opt/stayflow/scripts/collect-metrics.sh
METRICS=$(curl -s http://localhost:8080/health)
TIMESTAMP=$(date +%s)
echo "${TIMESTAMP} ${METRICS}" >> /var/log/stayflow-metrics.log

# Alert if errors spike (>10% error rate)
TOTAL=$(echo "$METRICS" | jq -r '.total_requests')
ERRORS=$(echo "$METRICS" | jq -r '.total_errors')
if [ "$TOTAL" -gt 0 ]; then
  RATE=$(echo "scale=2; $ERRORS * 100 / $TOTAL" | bc)
  if (( $(echo "$RATE > 10" | bc -l) )); then
    curl -s -X POST "https://your-webhook-url" \
      -d "StayFlow error rate: ${RATE}%"
  fi
fi
```

#### Option B: Prometheus + Grafana (if you want dashboards, ~100MB RAM)

```yaml
# Add to docker-compose.prod.yml when ready
  prometheus:
    image: prom/prometheus:latest
    container_name: stayflow-prometheus
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    ports:
      - "127.0.0.1:9090:9090"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: stayflow-grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
      GF_SERVER_ROOT_URL: https://metrics.yourdomain.com
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "127.0.0.1:3001:3000"
    restart: unless-stopped
```

### Disk Space Monitoring

```bash
# Cron: 0 * * * * (hourly)
USAGE=$(df / | tail -1 | awk '{print $5}' | tr -d '%')
if [ "$USAGE" -gt 85 ]; then
  curl -s -X POST "https://your-webhook-url" \
    -d "StayFlow VPS disk usage: ${USAGE}%"
fi
```

---

## 6. Security Checklist

### Infrastructure

| # | Check | Status | Notes |
|---|-------|--------|-------|
| 1 | SSH key-only authentication | ☐ | Disable `PasswordAuthentication` |
| 2 | UFW firewall (22, 80, 443 only) | ☐ | `ufw default deny incoming` |
| 3 | fail2ban enabled | ☐ | SSH brute-force protection |
| 4 | Automatic security updates | ☐ | `unattended-upgrades` |
| 5 | Docker socket not exposed | ☐ | Never bind to 0.0.0.0 |
| 6 | PostgreSQL bound to localhost only | ✅ | `127.0.0.1:5432:5432` |
| 7 | MinIO bound to localhost only | ✅ | `127.0.0.1:9000:9000` |
| 8 | Non-root container users | ✅ | `USER appuser` / `USER nextjs` |
| 9 | No `.env` in version control | ☐ | Add to `.gitignore` |
| 10 | Secrets min 32 chars in production | ✅ | Config validates this |

### Application

| # | Check | Status | Notes |
|---|-------|--------|-------|
| 11 | HTTPS enforced (HSTS) | ✅ | Caddy + `Strict-Transport-Security` |
| 12 | Security headers set | ✅ | `X-Frame-Options`, `X-Content-Type-Options` |
| 13 | CORS restricted to app domain | ✅ | `CORS_ALLOWED_ORIGINS` |
| 14 | Rate limiting on auth endpoints | ✅ | 5 req/min on `/auth` |
| 15 | Global rate limiting | ✅ | 100 req/min |
| 16 | Request body size limited | ✅ | 1MB default, 10MB for uploads |
| 17 | Graceful shutdown | ✅ | 30s timeout |
| 18 | SQL injection protection | ✅ | Parameterized queries via sqlc |
| 19 | JWT access ≠ refresh secret | ✅ | Config validates in production |
| 20 | Weak/dev secrets rejected | ✅ | Checks for "dev"/"change" substrings |

### Database

| # | Check | Status | Notes |
|---|-------|--------|-------|
| 21 | Strong password | ☐ | Generated, not default |
| 22 | Connection pooling configured | ✅ | 25 open, 10 idle, 5m lifetime |
| 23 | Backups automated | ✅ | Every 6 hours |
| 24 | Backup encryption at rest | ☐ | Add `gpg` to backup script |
| 25 | Restore tested | ☐ | Test monthly |

### Network

| # | Check | Status | Notes |
|---|-------|--------|-------|
| 26 | Internal services not publicly exposed | ✅ | Only Caddy on 80/443 |
| 27 | TLS 1.2+ only | ✅ | Caddy default |
| 28 | Server header removed | ✅ | `-Server` in Caddyfile |
| 29 | No debug endpoints in production | ☐ | Verify no `/debug/pprof` |

---

## 7. Findings & Recommendations

### Issues Found (Ordered by Severity)

#### 🔴 Critical

| # | Issue | Fix |
|---|-------|-----|
| 1 | **No backend health check in `docker-compose.prod.yml`** — Docker won't auto-restart on app hang | Add `healthcheck` to backend service |
| 2 | **No container resource limits** — OOM on 4GB VPS can kill all services | Add `deploy.resources.limits` |
| 3 | **No systemd unit** — Stack won't auto-start on VPS reboot | Create systemd service |

#### 🟡 Important

| # | Issue | Fix |
|---|-------|-----|
| 4 | **No Docker log rotation** — Disk fills over time | Add `daemon.json` log config |
| 5 | **MinIO data not backed up** — Document uploads lost on failure | Add `backup-minio.sh` |
| 6 | **`backup.sh` uses `stat -f%z`** — Only works on macOS; fails silently on Linux | Use `stat --format=%s` or `wc -c` |
| 7 | **No frontend health check in compose** — No restart on frontend crash | Add healthcheck to frontend |
| 8 | **Backup not encrypted** — Sensitive guest data in plaintext dumps | Add GPG encryption |

#### 🟢 Recommended

| # | Issue | Fix |
|---|-------|-----|
| 9 | No `.env.example` committed | Create documented template |
| 10 | No WAL archiving | Enable for point-in-time recovery |
| 11 | No monitoring beyond health checks | Add UptimeRobot + Healthchecks.io |
| 12 | No PostgreSQL connection timeout in healthcheck | Add `timeout` to health endpoint |

---

### Recommended Fixes to Apply

#### Fix 1: Add Backend Health Check to `docker-compose.prod.yml`

```yaml
  backend:
    # ... existing config ...
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/ready"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s
```

#### Fix 2: Add Resource Limits

```yaml
  backend:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'

  frontend:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'

  postgres:
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'

  minio:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'

  caddy:
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.25'
```

#### Fix 3: Systemd Service

Create `/etc/systemd/system/stayflow.service`:

```ini
[Unit]
Description=StayFlow Track Production
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/stayflow
ExecStart=/usr/bin/docker compose -f docker-compose.prod.yml up -d
ExecStop=/usr/bin/docker compose -f docker-compose.prod.yml down
TimeoutStartSec=120

[Install]
WantedBy=multi-user.target
```

Enable: `systemctl enable stayflow`

#### Fix 4: Docker Daemon Log Rotation

Create `/etc/docker/daemon.json`:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "5"
  }
}
```

---

## Quick-Start Deployment Script

```bash
#!/bin/bash
# One-time VPS setup script
# Run as root on fresh Ubuntu 24.04 VPS

set -euo pipefail

# System updates
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Create app user
useradd -m -s /bin/bash stayflow
usermod -aG docker stayflow

# Firewall
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# fail2ban
apt install -y fail2ban
systemctl enable fail2ban

# Swap (2GB)
fallocate -l 2G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# Auto security updates
apt install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# Docker log rotation
cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "5" }
}
EOF
systemctl restart docker

# Create app directory
mkdir -p /opt/stayflow/backups/{postgres,minio}
chown -R stayflow:stayflow /opt/stayflow

echo "VPS setup complete. Now:"
echo "1. Switch to stayflow user: su - stayflow"
echo "2. Clone repo to /opt/stayflow/"
echo "3. Create .env file"
echo "4. Run: docker compose -f docker-compose.prod.yml up -d"
```

---

## Uptime Budget (99.9%)

| Period | Allowed Downtime |
|--------|-----------------|
| Monthly | 43 minutes |
| Quarterly | 2 hours 10 min |
| Yearly | 8 hours 46 min |

With this setup:
- **Auto-restart** (Docker `restart: unless-stopped` + systemd) handles most crashes in < 30s
- **Health checks** detect hangs and trigger container restart
- **External monitoring** alerts you within 5 minutes of issues
- **Backup + restore** enables full recovery within 60 minutes
- **Caddy auto-SSL** eliminates certificate expiry downtime

This architecture achieves 99.9% for a solo-founder SaaS with no orchestration overhead.
