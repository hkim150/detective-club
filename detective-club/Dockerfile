# build binary
FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o detective-club .

# runtime image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app .

EXPOSE 80

ENTRYPOINT [ "./detective-club" ]