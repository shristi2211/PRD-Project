# ⛳ Golf Score Lottery (GSL)

A full-stack web application that combines golf score tracking with a charitable lottery system. Players submit their Stableford golf scores, which automatically enter them into monthly lottery draws. Winnings are split between the winner and their chosen charity.

---

## 🏗️ Tech Stack

| Layer      | Technology                                      |
| ---------- | ----------------------------------------------- |
| Frontend   | Vanilla JavaScript + Vite (SPA with Hash Router)|
| Backend    | Go 1.22 (Chi router, pgx, JWT RS256)            |
| Database   | PostgreSQL (Supabase)                            |
| Auth       | JWT RS256 (Access + Refresh tokens)              |
| Deployment | Frontend → Vercel · Backend → Render · DB → Supabase |

---

## 📁 Project Structure

```
PRD-Project/
├── frontend/               # Vite + Vanilla JS SPA
│   ├── index.html
│   ├── js/
│   │   ├── app.js          # Router & page mounting
│   │   ├── auth.js         # Authentication helpers
│   │   ├── pages/          # Page components (landing, login, register, dashboard, etc.)
│   │   └── utils/api.js    # Fetch wrapper with auto-refresh
│   ├── css/styles.css       # Full design system
│   ├── vercel.json          # Vercel SPA rewrite config
│   └── .env.example
│
├── backend/                # Golang REST API
│   ├── cmd/server/main.go  # Entry point & route wiring
│   ├── internal/
│   │   ├── config/         # Environment variable loader
│   │   ├── handler/        # HTTP handlers (auth, users, scores, draws, etc.)
│   │   ├── middleware/     # Auth, CORS, RBAC, rate limiting
│   │   ├── models/         # Request/response structs
│   │   ├── repository/     # Database queries (pgx)
│   │   ├── service/        # Business logic layer
│   │   └── utils/          # JWT, password hashing, key management
│   ├── migrations/         # SQL schema files (001–005)
│   ├── Dockerfile          # Multi-stage Docker build for Render
│   └── .env.example
│
└── README.md               # ← You are here
```

---

## 🚀 Deployment Guide

### 1. Database (New Supabase Project)

1. Go to [supabase.com](https://supabase.com) and create a **new project**.
2. Open the **SQL Editor** in the Supabase dashboard.
3. Run the migration files **in order**:
   - `backend/migrations/001_init.sql`
   - `backend/migrations/002_phase4_core.sql`
   - `backend/migrations/003_multi_charity.sql`
   - `backend/migrations/004_subscription_plans.sql`
   - `backend/migrations/005_ip_subscriptions.sql`
4. Copy the **Connection String (URI)** from:  
   `Settings → Database → Connection String → URI`  
   Append `?statement_cache_capacity=0&default_query_exec_mode=exec` to the URI.

### 2. Backend (Render.com)

1. Push your code to a **GitHub repository**.
2. On [render.com](https://render.com), create a new **Web Service**.
3. Connect your GitHub repo, set **Root Directory** to `backend`.
4. Set **Runtime** to `Docker`.
5. Add the following **Environment Variables**:

   | Variable              | Value                                             |
   | --------------------- | ------------------------------------------------- |
   | `DATABASE_URL`        | Your Supabase connection string (from Step 1.4)   |
   | `PORT`                | `8080`                                            |
   | `CORS_ORIGINS`        | `https://your-frontend.vercel.app` (update after Vercel deploy) |
   | `JWT_PRIVATE_KEY`     | Contents of `backend/keys/private.pem`            |
   | `JWT_PUBLIC_KEY`      | Contents of `backend/keys/public.pem`             |
   | `ACCESS_TOKEN_EXPIRY` | `15m`                                             |
   | `REFRESH_TOKEN_EXPIRY`| `168h`                                            |
   | `REDIS_URL`           | _(leave empty — rate limiting disabled)_          |

6. Deploy. Your Render URL is: `https://prd-project.onrender.com`.

### 3. Frontend (Vercel)

1. On [vercel.com](https://vercel.com), import the same GitHub repo.
2. Set **Root Directory** to `frontend`.
3. Set **Framework Preset** to `Vite`.
4. Add this **Environment Variable**:

   | Variable         | Value                                    |
   | ---------------- | ---------------------------------------- |
   | `VITE_API_URL`   | `https://prd-project.onrender.com`       |

5. Deploy. Your frontend URL will be something like `https://gsl-frontend.vercel.app`.

### 4. Post-Deploy: Update CORS

Go back to Render and update `CORS_ORIGINS` to your actual Vercel URL.

---

## 🔑 Test Credentials

After deploying, visit your Vercel URL and perform the initial Admin Setup:

### Admin Account (created via /admin-setup)
| Field    | Value                     |
| -------- | ------------------------- |
| Name     | `Admin User`              |
| Email    | `admin@gsl.com`           |
| Password | `Admin@123456`            |

### Test Subscriber (register via /register after selecting a plan)
| Field    | Value                     |
| -------- | ------------------------- |
| Name     | `Test User`               |
| Email    | `test@gsl.com`            |
| Password | `Test@123456`             |
| Plan     | `Monthly`                 |

---

## 🛠️ Local Development

### Prerequisites
- **Go 1.22+**
- **Node.js 18+** and npm
- **PostgreSQL** (or a Supabase project)

### Backend
```bash
cd backend
cp .env.example .env
# Edit .env with your DATABASE_URL
go run cmd/server/main.go
# Server starts on http://localhost:8080
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# Opens http://localhost:5173
```

---

## 📋 Features

### User Panel
- 📝 **Score Submission** — Enter Stableford scores (1–45) with round date
- 🏆 **Lottery Dashboard** — View monthly draws, winnings, and trends
- 💚 **Charity Selection** — Choose a charity and set contribution %
- 👤 **Profile Management** — Update profile, change password, delete account
- 📊 **Statistics** — Score trends, charity distribution charts

### Admin Panel
- 🎰 **Draw Management** — Run/simulate monthly lottery draws
- 👥 **User Management** — Search, filter, activate/deactivate users
- 📈 **Reports** — User, revenue, draw, and charity reports
- 🏢 **Charity Management** — CRUD operations on charities
- 🏅 **Winner Verification** — Approve/reject winner proof submissions
- 📜 **Activity Logs** — Drill-down user activity tracking
- 💳 **Subscription Tracking** — Monitor user subscription statuses

### Security
- 🔐 RS256 JWT authentication (access + refresh tokens)
- 🛡️ RBAC middleware (admin vs. user routes)
- 🔒 Subscription gate at login (free users blocked until subscribed)
- ⚡ Optional Redis-based rate limiting

---

## 📄 License

This project is for academic/evaluation purposes.
