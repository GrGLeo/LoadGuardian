FROM golang:1.23 AS builder

WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /app/main .
CMD ["./main"]
