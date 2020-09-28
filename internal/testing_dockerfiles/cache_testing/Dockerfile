FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/
RUN apk add build-base
COPY . .

ADD config.yaml /config.yaml

ENV CGO_ENABLED 1

RUN go test -c -v -race -tags=integration job-backend-trainee-assignment/internal/cache -o /go/bin/cache_test

ENV CGO_ENABLED 0

FROM alpine

COPY --from=builder /go/bin/cache_test /cache_test
COPY --from=builder /config.yaml /config.yaml

ENTRYPOINT ["/cache_test"]

