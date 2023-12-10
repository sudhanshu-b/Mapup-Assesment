FROM golang:latest

WORKDIR /app

COPY . .

RUN go build -o main .

EXPOSE 8000

CMD ["./main"]
