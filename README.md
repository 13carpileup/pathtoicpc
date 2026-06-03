# Path to ICPC

Basic full-stack scaffold with a Go backend and a React frontend using React Router.

## Project Structure

```text
backend/   Go HTTP API
frontend/  Vite + React single-page app
```

## Backend

```bash
cd backend
go run .
```

The API runs on `http://localhost:8080` by default.

Available endpoints:

- `GET /api/health`
- `GET /api/message`

Set a custom port with:

```bash
PORT=9000 go run .
```

## Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs on `http://localhost:5173` by default. During development, Vite proxies `/api` requests to the Go backend.

Routes:

- `/`
- `/about`
- `/dashboard`
