FROM "golang:1.12"

WORKDIR /app

COPY . /app

RUN go mod download

RUN go build -o main .

EXPOSE 8080

ENTRYPOINT ["./main", "docker"]