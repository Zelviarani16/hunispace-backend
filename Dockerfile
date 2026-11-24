# Tahap build
FROM golang:1.24-alpine AS build
WORKDIR /src
# go.mod dan go.sum disalin lebih dulu agar layer dependency di-cache
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

# Tahap jalan (image kecil). /app hanya berisi: app dan logs
FROM alpine:3.20
RUN adduser -D -u 10001 appuser \
 && mkdir -p /app/logs \
 && chown -R appuser:appuser /app
WORKDIR /app
COPY --from=build /out/app /app/app
USER appuser
EXPOSE 3000
ENV APP_PORT=3000
CMD ["/app/app"]
