FROM golang

WORKDIR $GOPATH/src/app
COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

CMD [ "app" ]