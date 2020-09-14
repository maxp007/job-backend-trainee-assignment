FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/
RUN apk add build-base
COPY . .

ENV CGO_ENABLED 1

RUN go test -c -v -race -tags=integration job-backend-trainee-assignment/internal/app -o /go/bin/app_test

ENV CGO_ENABLED 0

FROM alpine

COPY --from=builder /go/bin/app_test /test/app_test

ENTRYPOINT ["/test/app_test"]

