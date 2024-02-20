FROM golang:1.22-alpine as builder


WORKDIR $GOTPATH/src/github.com/antonydenyer/block-builder-mempool
COPY . .

RUN CGO_ENABLED=0 go mod download

RUN CGO_ENABLED=0 go build -o /app ./cmd/app

FROM alpine:3.19

RUN addgroup --system --gid 1000 appgroup \
    && adduser --system --uid 1000 appuser -G appgroup

USER appuser

COPY --from=builder /app /usr/bin/app

EXPOSE 8080

ENTRYPOINT [ "app" ]
