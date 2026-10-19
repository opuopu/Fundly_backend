# Fundly API Production Deployment Guide

This guide covers deploying the Fundly Go backend with **zero cost** and **no credit card required**.

---

## ⚡ Option 1: Vercel + Neon PostgreSQL (100% Free, No Credit Card Required)

Vercel provides free serverless hosting without asking for card details, and Neon provides a free managed PostgreSQL cloud database.

### Step 1: Create a Free Database on Neon (Takes 60 seconds)
1. Go to [neon.tech](https://neon.tech) and click **Sign Up** (Choose "Continue with GitHub").
2. Create a new project (name it `fundly`).
3. Neon will display your connection string:
   ```text
   postgresql://fundly_owner:xxxx@ep-xyz.us-east-2.aws.neon.tech/fundly?sslmode=require
   ```
4. Copy this string.

### Step 2: Deploy to Vercel via CLI
Since you are already logged in to the Vercel CLI (`allendodul6-9229`), run:
```bash
vercel
```
When prompted:
- **Set up and deploy?** `Y`
- **Which scope?** Choose your account (`mozammel-hoque-doduls-projects`)
- **Link to existing project?** `N`
- **Project name?** `fundly-api`
- **In which directory is your code located?** `./`
- **Want to modify settings?** `N`

### Step 3: Set Environment Variables on Vercel
Run these commands to add your environment variables to Vercel:
```bash
vercel env add DATABASE_URL production
# Paste your Neon connection string when prompted

vercel env add AUTO_MIGRATE production
# Enter: true

vercel env add JWT_SECRET production
# Enter: a-secret-key-with-at-least-32-characters

vercel env add SMTP_HOST production
# Enter: smtp.gmail.com

vercel env add SMTP_PORT production
# Enter: 587

vercel env add SMTP_USERNAME production
# Enter: allendodul6@gmail.com

vercel env add SMTP_PASSWORD production
# Enter: ndikmjvgdzbelwsk

vercel env add SMTP_FROM production
# Enter: allendodul6@gmail.com
```

### Step 4: Deploy to Production
```bash
vercel --prod
```
Your Fundly API will be live on `https://fundly-api.vercel.app`!
- Database migrations will apply automatically on the first request (`AUTO_MIGRATE=true`).
- Test your deployment: `curl https://fundly-api.vercel.app/health`

---

## 🌐 Option 2: Expose Your Local Server via Cloudflare Tunnel (100% Free, No Card)

Since your local Go server and Docker PostgreSQL are already working 100%, you can expose them to the public internet securely without opening router ports or paying anything:

```bash
# Install cloudflared (Ubuntu/Debian)
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb

# Run tunnel pointing to local port 8081
cloudflared tunnel --url http://localhost:8081
```
Cloudflare will give you an instant public HTTPS URL like:
`https://your-random-subdomain.trycloudflare.com`

---

## 🐳 Option 3: Self-Host via Docker Compose (VPS / EC2 / Hetzner)

```bash
git clone https://github.com/Dodul01/fundly.git
cd fundly
cp .env.example .env
nano .env
docker compose up -d --build
```
