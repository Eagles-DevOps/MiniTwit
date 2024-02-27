FROM golang:1.18

WORKDIR ./minitwit-web-app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY *.go go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app ./...

CMD ["./app"]