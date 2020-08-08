FROM golang:1.10

WORKDIR $GOPATH/src/github.com/arazmj/GoCache

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

EXPOSE 8080

CMD ["GoCache"]