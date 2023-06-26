FROM golang as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
FROM alpine:latest
COPY --from=builder /app/main /app/main
WORKDIR /app
EXPOSE 8080
CMD ["./main"]
