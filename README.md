# Path to ICPC

Codeforces does not provide good methods to train. You have regular contests (which take place at 7am Vancouver time...), virtual contests, and the problemset. If you want a timed practice experience that lasts less than two hours, you're out of luck. The main purpose of this project is to create a practice environment that is a little more flexible, allowing for the completion of individual problems under timed conditions.

## Methodology

Most codeforces problems have an associated rating. A user with a rating of $r$ has a greater than 50% chance of solving any problem with rating $r_1\le r$ within its contest period. We extrapolate and bastardise this fact to calculate our own ratings without needing to sit full contests.

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
- `GET /api/user.info?handles=tourist`
- `GET /api/user.status?handle=tourist`
- `GET /api/problemset.problems`

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
