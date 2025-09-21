FROM golang

WORKDIR /api

COPY . .

RUN go mod download

CMD ["go", "run", "./cmd/api/main.go"]
