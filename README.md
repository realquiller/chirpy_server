# Chirpy Server 🐦

A backend API for **Chirpy**, a minimalist Twitter clone built with Go. This project was developed as part of the Boot.dev backend course, focusing on RESTful API design, authentication, and database interactions.

This project includes JWT-based authentication, refresh token support, secure webhook handling, and CRUD operations for user-generated chirps.

> ⚠️ Educational project — not production-ready.  
> API keys and secrets in this repo are random or dummy.

---

## 🚀 Features

- **User Authentication**: Secure user registration and login with JWT-based authentication.
- **Chirp Management**: Create, retrieve, and delete chirps (short messages).
- **Chirpy Red Membership**: Upgrade users to premium status via webhook integration.
- **API Key Verification**: Secure webhook endpoints using API keys.
- **Metrics Tracking**: Monitor API usage with built-in metrics.
- **Admin Controls**: Reset and manage application data through admin endpoints.

---

## 🛠️ Tech Stack

- **Go**: Core language for server-side development.
- **PostgreSQL**: Relational database for storing users and chirps.
- **Goose**: Database migration tool.
- **UUID**: Unique identifiers for users and chirps.
- **JWT**: Authentication tokens for secure API access.

## 🔑 Environment Variables

| Variable     | Description                            |
|--------------|----------------------------------------|
| `DB_URL`     | PostgreSQL connection string           |
| `SECRET`     | Secret key for signing JWT tokens      |
| `POLKA_KEY`  | Secret key for authenticating webhooks |
| `PLATFORM`   | Used for allowing dev-only features    |

Example `.env`:

```env
DB_URL=postgres://postgres:postgres@localhost:5432/chirpy
SECRET=randomly-generated-dev-secret
POLKA_KEY=f271c81ff7084ee5b99a5091b42d486e
PLATFORM=dev
```

# 🛠️ Getting Started

## Clone the repo
``` bash
git clone https://github.com/realquiller/chirpy_server.git
cd chirpy_server
```

## Set up environment
``` bash
cp .env.example .env
```

## Run database migrations
``` bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose postgres "$DB_URL" up
```

## Build and run the app
``` bash
go build -o out && ./out
```

# 📡 Endpoints Overview
| Method | Route                       | Description                              |
|--------|-----------------------------|------------------------------------------|
| GET    | `/api/healthz`              | Health check                             |
| POST   | `/api/users`                | Register new user                        |
| POST   | `/api/login`                | Login and get JWT & refresh token        |
| PUT    | `/api/users`                | Update email and password (auth required)|
| GET    | `/api/chirps`               | Get all chirps (filter & sort optional)  |
| GET    | `/api/chirps/{chirpid}`     | Get specific chirp by ID                 |
| POST   | `/api/chirps`               | Create chirp (auth required)             |
| DELETE | `/api/chirps/{chirpid}`     | Delete chirp (author only)               |
| POST   | `/api/refresh`              | Get new access token via refresh token   |
| POST   | `/api/revoke`               | Revoke refresh token                     |
| POST   | `/api/polka/webhooks`       | Handle Chirpy Red upgrade (via Polka)    |

# 🎯 Project Goals

This project helped me practice:

- API structuring in Go  
- Secure handling of JWT + refresh tokens  
- Database migrations with Goose  
- Writing and validating SQL queries  
- Building idempotent webhooks  
- Managing edge cases, 401s, 403s, etc.

# 🧠 Lessons Learned

- 🔐 Don’t trust incoming requests — always verify.  
- 🚦 Use status codes precisely: 204, 403, 401, 404 all have meaning.  
- 🤖 Webhooks are just requests… but sneakier.  
- 🧹 It pays off to keep handlers clean and structured.  
- 👀 Readability matters — especially when debugging with sleep-deprived eyes.

# License

This is a learning project and has no license. Feel free to peek and learn.  
If you wanna collab on cool Go stuff, hit me up 😎

