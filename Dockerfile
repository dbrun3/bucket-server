FROM golang:1.25

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bs

EXPOSE 8080

CMD ["/bs"]