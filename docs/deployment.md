# Fundly API Production Deployment Guide

This guide covers deploying the Fundly Go backend with PostgreSQL, persistent storage for attachments/avatars, and automated migrations.

---

## 🚀 Option 1: Deploy to Render (Recommended)

Render provides native Docker container hosting, a managed PostgreSQL database, and persistent SSD disks for file uploads.

### Step 1: Push Code to GitHub
Ensure your repository contains [`render.yaml`](../render.yaml) and [`Dockerfile`](../Dockerfile).

### Step 2: Create a Blueprint on Render
1. Go to your [Render Dashboard](https://dashboard.render.com/).
2. Click **New +** → **Blueprint**.
3. Connect your GitHub repository (`fundly`).
4. Render will automatically detect [`render.yaml`](../render.yaml) and configure:
   - **PostgreSQL Database** (`fundly-db`): Automatically linked via `DATABASE_URL`.
   - **Web Service** (`fundly-api`): Built from the multi-stage [`Dockerfile`](../Dockerfile).
   - **Persistent Disk** (`fundly-uploads`): Mounted at `/app/uploads` (1GB SSD) so avatars and receipts persist across redeploys.

### Step 3: Set Secret Environment Variables
In the Render Web Service Settings -> **Environment Variables**, fill in:
- `SMTP_USERNAME`: `allendodul6@gmail.com`
- `SMTP_PASSWORD`: `ndikmjvgdzbelwsk`
- `SMTP_FROM`: `allendodul6@gmail.com`
- `JWT_SECRET`: (Render auto-generates this, or you can paste your own 32+ character random string)

### Step 4: Deploy
Click **Apply**. Render will:
1. Provision the PostgreSQL database.
2. Build the Docker container.
3. Automatically execute embedded database migrations on startup (`AUTO_MIGRATE=true`).
4. Expose the API on your Render URL (e.g. `https://fundly-api.onrender.com`).
5. Monitor health via `GET /health`.

---

## 🚆 Option 2: Deploy to Railway

Railway runs containers natively and provisions managed PostgreSQL in seconds.

### Step 1: Initialize Project on Railway
1. Go to [railway.app](https://railway.app) and click **New Project**.
2. Select **Provision PostgreSQL** to spin up a managed database.
3. In the same project, click **New** → **GitHub Repo** → select `fundly`.

### Step 2: Configure Service Variables
Under your `fundly` service → **Variables**:
- Set `DATABASE_URL` = `${{Postgres.DATABASE_URL}}` (Railway references the database connection string automatically).
- Set `AUTO_MIGRATE` = `true`.
- Set `APP_ENV` = `production`.
- Set `JWT_SECRET` = `<your-jwt-secret>`.
- Set `SMTP_HOST` = `smtp.gmail.com`.
- Set `SMTP_PORT` = `587`.
- Set `SMTP_USERNAME` = `allendodul6@gmail.com`.
- Set `SMTP_PASSWORD` = `ndikmjvgdzbelwsk`.
- Set `SMTP_FROM` = `allendodul6@gmail.com`.

### Step 3: Add Volume for File Uploads
1. Go to your service settings in Railway.
2. Under **Volumes**, click **Add Volume**.
3. Set the Mount Path to `/app/uploads`.

Railway will build using [`Dockerfile`](../Dockerfile) and start your server with live health checks.

---

## 🐳 Option 3: Self-Host via Docker Compose (VPS / EC2 / Hetzner)

If running on a private VPS (Ubuntu/Debian):

```bash
# Clone the repository
git clone https://github.com/Dodul01/fundly.git
cd fundly

# Copy and edit production environment variables
cp .env.example .env
nano .env

# Build and start services in the background
docker compose up -d --build
```

---

## ℹ️ Why Vercel is Not Recommended for this Backend

Vercel is built for frontend apps and serverless micro-functions (AWS Lambda under the hood). Running a full Go API like Fundly on Vercel introduces severe architectural limitations:
1. **No Local File Storage:** Vercel functions have an ephemeral, read-only filesystem (`./uploads` cannot persist files between requests).
2. **No Persistent Process:** Long-running server processes, background workers, and in-memory token blacklists are wiped after every invocation.
3. **Database Connectivity:** Vercel serverless functions cannot connect to local Docker containers (`localhost`) and require an external cloud database (Neon, Supabase) with connection poolers.
