# --- Build aşaması ---
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Önce sadece go.mod ve go.sum'u kopyalayıp bağımlılıkları indiriyoruz.
# Bu, kod değişse bile bağımlılıklar değişmediyse Docker'ın bu adımı
# tekrar tekrar yapmaması için önemli bir performans tekniği (layer caching).
COPY go.mod go.sum ./
RUN go mod download

# Şimdi tüm kaynak kodu kopyalıyoruz
COPY services/engine-service/ .

# Statik bir binary derliyoruz (CGO_ENABLED=0, alpine gibi minimal
# imajlarda çalışabilmesi için önemli)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/engine-service .

# --- Çalıştırma aşaması ---
FROM alpine:latest

WORKDIR /app

# Sadece derlenmiş binary'yi build aşamasından kopyalıyoruz.
# Go compiler'ı, kaynak kodu, hiçbir şey gerekmiyor — final imaj çok küçük olur.
COPY --from=builder /app/engine-service .

EXPOSE 8080

CMD ["./engine-service"]