FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get -u github.com/lib/pq

CMD ["go", "run", "./cmd/web"]
