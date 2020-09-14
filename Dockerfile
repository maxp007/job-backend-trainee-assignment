FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/

COPY . .
COPY ./config.json /go/bin/config.json


ENV CGO_ENABLED 0

RUN go build -o /go/bin/bill_service

FROM scratch

COPY --from=builder /go/bin/config.json /app/config.json
COPY --from=builder /go/bin/bill_service /app/bill_service

ENTRYPOINT ["/app/bill_service"]
