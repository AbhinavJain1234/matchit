# MatchIt

Monorepo for the MatchIt ride-matching project.

## Repository Layout

- backend: Go backend APIs and matching logic
- mobile: Flutter mobile apps (driver/rider)
- web: Web frontend/admin dashboard

## Run Backend

1. cd backend
2. go mod tidy
3. go run cmd/server/main.go

Backend expects environment variables in backend/.env.

## Next Steps

- Initialize Flutter app in mobile
- Initialize web app in web
- Add docker-compose and infra in a later phase
