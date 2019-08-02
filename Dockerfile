# FROM golang:latest

# WORKDIR /app

# COPY . .

# RUN go get github.com/gorilla/mux
# RUN go build -o main .

FROM golang:latest
ADD . /go/src/testGorillaMux
WORKDIR /go/src/testGorillaMux
RUN go get testGorillaMux
RUN go install

EXPOSE 8080

CMD ["/go/bin/testGorillaMux", "config.json"]