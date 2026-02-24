FROM golang:latest
WORKDIR /app
COPY main.go .
RUN go mod init producer && go get github.com/rabbitmq/amqp091-go
RUN go build -o /producer main.go
CMD ["/producer"]