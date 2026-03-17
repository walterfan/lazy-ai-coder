# Docker Deployment Guide

Deploy lazy-ai-coder with all dependencies using Docker Compose.

## Quick Start

### 1. Configure Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit with your credentials
nano .env
```

**Required variables:**
```bash
# Database
DB_USER=postgres
DB_PASS=your_secure_db_password
DB_NAME=lazy_ai_coder

# Redis (NEW - Required for security)
REDIS_PASSWORD=your_secure_redis_password

# LLM API
LLM_API_KEY=sk-your-api-key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4

# GitLab
GITLAB_TOKEN=glpat-your-token
GITLAB_BASE_URL=https://gitlab.com
```

### 2. Start Services

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f lazy-ai-coder
```

### 3. Access Services

| Service | URL | Description |
|---------|-----|-------------|
| **lazy-ai-coder** | http://localhost:8888 | Main application (web + MCP) |
| **JupyterLab** | http://localhost:8889 | Interactive Python notebooks |
| **PlantUML** | http://localhost:8000 | Diagram generation |
| **PgWeb** | http://localhost:8081 | PostgreSQL web UI |
| **Redis** | localhost:6379 | Redis cache (password protected) |
| **PostgreSQL** | localhost:5432 | PostgreSQL database |

## Architecture

```
┌─────────────────────────────────────────┐
│           Nginx (80/443)                │
│         Reverse Proxy + SSL             │
└─────────────┬───────────────────────────┘
              │
    ┌─────────┼──────────┬────────────┐
    │         │          │            │
┌───▼───┐ ┌──▼────┐ ┌───▼─────┐ ┌────▼────┐
│ App   │ │PlantU-│ │ PgWeb   │ │ Redis   │
│ :8888 │ │ML     │ │ :8081   │ │ :6379   │
│       │ │ :8080 │ │         │ │         │
└───┬───┘ └───────┘ └────┬────┘ └─────────┘
    │                    │
    └────────┬───────────┘
             │
      ┌──────▼──────┐
      │ PostgreSQL  │
      │   :5432     │
      │  (pgvector) │
      └─────────────┘
```

## Services

### 1. lazy-ai-coder (Port 8888)

Main application providing:
- Web UI for LLM interactions
- HTTP MCP server
- GitLab integration
- Code analysis tools

**Environment variables:**
```yaml
LLM_API_KEY: Your LLM API key
LLM_BASE_URL: LLM endpoint
LLM_MODEL: Model to use
GITLAB_TOKEN: GitLab access token
PLANTUML_URL: PlantUML service URL
```

### 2. PostgreSQL with pgvector (Port 5432)

Database with vector extension for:
- User data storage
- Vector embeddings (if used)
- Session management

**Security:** Password protected via `DB_PASS`

### 3. Redis (Port 6379) 

**🔐 NOW PASSWORD PROTECTED!**

Used for:
- Session caching
- Rate limiting
- Temporary data storage

**Security:** 
- Requires password authentication (`REDIS_PASSWORD`)
- No anonymous access allowed

**Connect with password:**
```bash
# CLI
redis-cli -h localhost -p 6379 -a your_redis_password

# Python
import redis
r = redis.Redis(host='localhost', port=6379, password='your_redis_password')
```

### 4. PlantUML Server (Port 8000)

Generates diagrams from PlantUML scripts:
- Sequence diagrams
- Class diagrams
- Mind maps
- Component diagrams

### 5. PgWeb (Port 8081)

Web interface for PostgreSQL database management.

### 6. JupyterLab (Port 8889)

**🔐 PASSWORD PROTECTED!**

Interactive Python development environment:
- Python notebooks for data analysis
- Scientific computing libraries (NumPy, Pandas, SciPy)
- Direct access to PostgreSQL and Redis
- Shared notebook storage

**Security:**
- Requires password authentication (`JUPYTER_PASSWORD`)
- Token-based access
- Isolated container environment

**Access:**
```
URL: http://localhost:8889
Password: Your JUPYTER_PASSWORD from .env
```

**Pre-installed packages:**
- NumPy, Pandas, Matplotlib, SciPy
- Scikit-learn, Seaborn
- IPython, JupyterLab extensions

**Connect to services from notebooks:**
```python
# PostgreSQL
import psycopg2
conn = psycopg2.connect(
    host="pgvector",
    database="lazy_ai_coder",
    user="postgres",
    password="your_db_password"
)

# Redis
import redis
r = redis.Redis(
    host='redis',
    port=6379,
    password='your_redis_password'
)

# Test MCP HTTP endpoint
import requests
response = requests.get('http://lazy-ai-coder:8888/api/v1/mcp/info')
print(response.json())
```

### 7. Nginx (Ports 80/443)

Reverse proxy providing:
- SSL termination
- Load balancing
- Request routing

## Configuration

### Redis Password Configuration

**⚠️ IMPORTANT:** Redis is now password-protected for security.

1. **Set password in .env:**
```bash
REDIS_PASSWORD=your_strong_random_password_here
```

2. **Generate strong password:**
```bash
# Linux/macOS
openssl rand -base64 32

# Or use
pwgen 32 1
```

3. **Connect to Redis:**
```bash
# Using redis-cli
docker exec -it redis redis-cli -a your_redis_password

# Test connection
redis-cli -h localhost -p 6379 -a your_redis_password ping
# Should return: PONG
```

### SSL Configuration (Optional)

Generate SSL certificates:

```bash
cd nginx
./generate-ssl.sh
```

Or use Let's Encrypt:

```bash
# Install certbot
apt-get install certbot

# Generate certificate
certbot certonly --standalone -d yourdomain.com

# Copy to nginx/ssl/
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/
```

### Database Schema

Initialize database schema:

```bash
# PostgreSQL
docker exec -i pgvector psql -U postgres -d lazy_ai_coder < db/schema_pg.sql

# Or MySQL
docker exec -i mysql mysql -u root -p lazy_ai_coder < db/schema_mysql.sql
```

## Management Commands

### Start/Stop Services

```bash
# Start all
docker-compose up -d

# Stop all
docker-compose down

# Restart specific service
docker-compose restart lazy-ai-coder

# Rebuild and restart
docker-compose up -d --build lazy-ai-coder
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f lazy-ai-coder
docker-compose logs -f redis
docker-compose logs -f pgvector

# Last 100 lines
docker-compose logs --tail=100 lazy-ai-coder
```

### Database Management

```bash
# Backup database
docker exec pgvector pg_dump -U postgres lazy_ai_coder > backup.sql

# Restore database
docker exec -i pgvector psql -U postgres lazy_ai_coder < backup.sql

# Connect to database
docker exec -it pgvector psql -U postgres -d lazy_ai_coder
```

### Redis Management

```bash
# Connect to Redis CLI (with password)
docker exec -it redis redis-cli -a your_redis_password

# Check Redis info
docker exec redis redis-cli -a your_redis_password INFO

# Monitor Redis commands
docker exec redis redis-cli -a your_redis_password MONITOR

# Flush all data (DANGER!)
docker exec redis redis-cli -a your_redis_password FLUSHALL
```

### Health Checks

```bash
# Check all services status
docker-compose ps

# Test lazy-ai-coder
curl http://localhost:8888/api/v1/mcp/info

# Test JupyterLab (should redirect to login)
curl -I http://localhost:8889

# Test PlantUML
curl http://localhost:8000/png/SyfFKj2rKt3CoKnELR1Io4ZDoSa70000

# Test Redis (with password)
redis-cli -h localhost -p 6379 -a your_redis_password PING

# Test PostgreSQL
docker exec pgvector pg_isready -U postgres
```

## Troubleshooting

### Redis Connection Refused

**Problem:** Services can't connect to Redis

**Solution:** Check password is set correctly:
```bash
# Verify password in .env
grep REDIS_PASSWORD .env

# Test connection
docker exec redis redis-cli -a your_redis_password PING
```

### Port Conflicts

**Problem:** Port already in use

**Solution:** Change port mapping in docker-compose.yaml:
```yaml
ports:
  - "8889:8888"  # Changed from 8888:8888
```

### Permission Issues

**Problem:** Volume mount permission denied

**Solution:**
```bash
# Fix ownership
sudo chown -R $(whoami):$(whoami) ./db ./images

# Or run with correct user
docker-compose down
docker-compose up -d
```

### Build Failures

**Problem:** Docker build fails

**Solution:**
```bash
# Clean build
docker-compose down -v
docker-compose build --no-cache lazy-ai-coder
docker-compose up -d
```

### Redis Authentication Errors

**Problem:** "NOAUTH Authentication required"

**Solution:** Update application config to use password:
```bash
# Check lazy-ai-coder has Redis password
# Add to docker-compose.yaml if needed:
environment:
  - REDIS_PASSWORD=${REDIS_PASSWORD}
  - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
```

## Security Best Practices

### 1. Use Strong Passwords

```bash
# Generate secure passwords
openssl rand -base64 32  # For Redis
openssl rand -base64 32  # For Database
openssl rand -base64 32  # For Admin
```

### 2. Limit Network Exposure

```yaml
# Don't expose ports publicly
expose:
  - "6379"  # Instead of ports: - "6379:6379"
```

### 3. Use SSL/TLS

- Enable HTTPS in Nginx
- Use SSL for database connections
- Set `sslmode=require` in DB_URL

### 4. Regular Updates

```bash
# Update images
docker-compose pull
docker-compose up -d
```

### 5. Backup Strategy

```bash
# Automated backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
docker exec pgvector pg_dump -U postgres lazy_ai_coder > backup_$DATE.sql
docker exec redis redis-cli -a $REDIS_PASSWORD SAVE
```

## Production Deployment

### 1. Use Production Configuration

```bash
# production.env
REDIS_PASSWORD=<strong-random-password>
DB_PASS=<strong-random-password>
ADMIN_PASSWORD=<strong-random-password>
```

### 2. Enable SSL

- Configure Let's Encrypt
- Update nginx.conf for HTTPS
- Redirect HTTP to HTTPS

### 3. Set Resource Limits

```yaml
services:
  lazy-ai-coder:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

### 4. Enable Monitoring

- Add Prometheus exporter
- Configure log aggregation
- Set up health check endpoints

### 5. Configure Backup

- Automated daily backups
- Off-site backup storage
- Test restore procedures

## Scaling

### Horizontal Scaling

```bash
# Scale lazy-ai-coder
docker-compose up -d --scale lazy-ai-coder=3
```

### Load Balancing

Update nginx.conf:
```nginx
upstream app {
    server lazy-ai-coder-1:8888;
    server lazy-ai-coder-2:8888;
    server lazy-ai-coder-3:8888;
}
```

## Monitoring

### Resource Usage

```bash
# Container stats
docker stats

# Specific container
docker stats lazy-ai-coder redis pgvector
```

### Health Endpoints

```bash
# Application health
curl http://localhost:8888/api/v1/mcp/info

# Database health
docker exec pgvector pg_isready

# Redis health
docker exec redis redis-cli -a your_redis_password PING
```

## Support

- **Logs**: `docker-compose logs -f`
- **Documentation**: See main README.md
- **Issues**: Check application logs in container

---

**Ready to deploy!** 🚀

```bash
cp .env.example .env
# Edit .env with your credentials
docker-compose up -d
```
