FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/
RUN apk add build-base
COPY . .

ENV CGO_ENABLED 1

RUN go test -c -v -race -tags=integration job-backend-trainee-assignment/internal/http_app_handler -o /go/bin/http_app_handler_test

ENV CGO_ENABLED 0

FROM scratch

COPY --from=builder /go/bin/http_app_handler_test /test/http_app_handler_test

ENTRYPOINT ["/test/http_app_handler_test"]

