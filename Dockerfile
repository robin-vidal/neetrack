FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./internal/httpapi/assets/dist
RUN go build -trimpath -ldflags="-s -w" -o /out/neetrack ./cmd/neetrack

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=go-builder /out/neetrack /neetrack
USER nonroot:nonroot
ENTRYPOINT ["/neetrack"]
