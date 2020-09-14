FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/

COPY . .

ENV CGO_ENABLED 0

RUN go build  -o /go/bin/bill_service

FROM alpine

COPY --from=builder /go/bin/bill_service /app/bill_service

ENTRYPOINT ["/app/bill_service"]
