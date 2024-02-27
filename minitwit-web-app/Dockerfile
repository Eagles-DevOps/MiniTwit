FROM golang:1.18

WORKDIR ./

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY ./minitwit-web-app/*.go ./minitwit-web-app/go.mod ./minitwit-web-app/go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o app ./...

CMD ["./app"]