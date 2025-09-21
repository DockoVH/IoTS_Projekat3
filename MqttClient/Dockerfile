FROM golang

WORKDIR /api

COPY . .

RUN go mod download

EXPOSE 8085

CMD ["go", "run", "./cmd/api/main.go"]
