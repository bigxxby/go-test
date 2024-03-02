FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get -u github.com/lib/pq
RUN go get -u github.com/go-redis/redis

CMD ["go", "run", "./cmd/web"]
