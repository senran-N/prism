FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /prism ./cmd/prism/

FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /prism .
COPY --from=frontend /web/.next/standalone ./web/
COPY --from=frontend /web/.next/static ./web/.next/static
COPY --from=frontend /web/public ./web/public
COPY migrations/ migrations/
EXPOSE 8080 3000
CMD ["./prism"]
