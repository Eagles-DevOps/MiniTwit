FROM golang:1.22-alpine3.19 AS Build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY *.go go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /minitwit/app .


FROM alpine:latest
RUN apk add --no-cache bash
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh wait-for-it.sh
RUN chmod +x wait-for-it.sh

COPY --from=Build /minitwit/app /minitwit/app


CMD ["/minitwit/app"]