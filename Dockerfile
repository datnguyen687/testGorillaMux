FROM golang:latest as builder

ADD . /go/src/testGorillaMux

WORKDIR /go/src/testGorillaMux

RUN go get testGorillaMux

RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/bin/testGorillaMux .
COPY --from=builder /go/src/testGorillaMux/config.json .

EXPOSE 8080

CMD ["./testGorillaMux", "config.json"]