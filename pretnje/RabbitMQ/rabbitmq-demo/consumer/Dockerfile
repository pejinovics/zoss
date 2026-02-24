FROM golang:latest
WORKDIR /app
COPY main.go .
RUN go mod init consumer && go get github.com/rabbitmq/amqp091-go github.com/minio/minio-go/v7
RUN go build -o /consumer main.go
CMD ["/consumer"]