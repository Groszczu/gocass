FROM golang:1.19.5-alpine

WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./internal ./internal
COPY ./cmd/main.go ./

RUN go build -o /gocass

CMD ["/gocass"]

