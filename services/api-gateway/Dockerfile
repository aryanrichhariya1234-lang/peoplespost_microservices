# ================= BUILD =================
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 🔥 build static binary (fixes exec error)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

# ================= RUN =================
FROM alpine:latest

WORKDIR /root/

# required for HTTPS calls (Gemini, Cloudinary)
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/main ./main

RUN chmod +x ./main

EXPOSE 4000

CMD ["./main"]