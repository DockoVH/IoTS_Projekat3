FROM golang

WORKDIR /api

COPY . .

RUN go mod download

EXPOSE 5050

CMD ["go", "run", "./cmd/api/main.go"]
