# Payments Hexagonal - (c) 2026 Example Org
FROM golang:1.19

WORKDIR /app

COPY . .
RUN go mod tidy

EXPOSE 3000

CMD ["sh", "-c", "go run ./..."]
