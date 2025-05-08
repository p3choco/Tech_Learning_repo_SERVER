# backend/Dockerfile
FROM golang:1.24.2-alpine

WORKDIR /app

# --- dodajemy git i certyfikaty ---
RUN apk add --no-cache git ca-certificates gcc musl-dev sqlite-dev

ENV CGO_ENABLED=1

# kopiujemy definicje modułów i pobieramy zależności
COPY go.mod go.sum ./
RUN go mod download

# kopiujemy resztę kodu i budujemy
COPY . .
RUN go build -o server main.go

EXPOSE 8080
CMD ["./server"]
