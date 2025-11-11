FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o bot .


FROM alpine:3.22

WORKDIR /app
ENV TZ=Europe/Berlin
ENV WORK_DIR=/app/data
RUN mkdir -p ${WORK_DIR} && adduser -D botuser && chown -R botuser:botuser /app
COPY --from=builder /app/bot ./HereBeDragonsBot
USER botuser

VOLUME ["/app/data"]
CMD ["./HereBeDragonsBot"]