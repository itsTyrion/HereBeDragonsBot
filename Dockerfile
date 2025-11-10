FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o bot .

# -----

FROM alpine:3.22

WORKDIR /app

RUN mkdir -p /app/data && \
    addgroup -S botuser && \
    adduser -S botuser -G botuser && \
    chown -R botuser:botuser /app

COPY --from=builder /app/bot .

USER botuser

ENV WORK_DIR=/app/data

RUN mkdir -p ${WORK_DIR}

VOLUME ["/app/data"]

ENTRYPOINT ["./bot"]