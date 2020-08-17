FROM golang:1.15

WORKDIR $GOPATH/src/github.com/arazmj/gerdu

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

EXPOSE 8080

CMD ["gerdu"]