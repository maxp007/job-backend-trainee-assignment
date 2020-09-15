FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/

COPY . .

ADD config.yaml /config.yaml

ENV CGO_ENABLED 0

RUN go build  -o /go/bin/bill_service

FROM scratch

COPY --from=builder /go/bin/bill_service /bill_service
COPY --from=builder /config.yaml /config.yaml

ENTRYPOINT ["/bill_service"]
