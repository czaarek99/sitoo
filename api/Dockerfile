FROM "golang:1.13"

WORKDIR /app

COPY /api /app/api

RUN go mod download

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]