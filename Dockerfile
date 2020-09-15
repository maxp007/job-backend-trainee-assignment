FROM golang:1.14-alpine AS builder

WORKDIR $GOPATH/src/job-backend-trainee-assignent/

COPY . .

ADD config.yaml /config.yaml

ENV CGO_ENABLED 0

RUN go build  -o /go/bin/bill_service

FROM scratch

RUN apk update
RUN apk upgrade
RUN apk add ca-certificates && update-ca-certificates

RUN apk add --update tzdata
ENV TZ=Europe/Moscow
RUN rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/bill_service /bill_service
COPY --from=builder /config.yaml /config.yaml

ENTRYPOINT ["/bill_service"]
