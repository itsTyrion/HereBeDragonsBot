FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bot .

FROM alpine:3.24

WORKDIR /app
ENV TZ=Europe/Berlin
ENV WORK_DIR=/app/data
RUN mkdir -p ${WORK_DIR} && adduser -D botuser && chown -R botuser:botuser /app
COPY --from=builder /app/bot ./HereBeDragonsBot
USER botuser

VOLUME ["/app/data"]
CMD ["./HereBeDragonsBot"]