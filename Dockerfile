FROM golang:1.9

WORKDIR /go/src/app
COPY ./src/crawler .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."
RUN mkdir /crawls

CMD ["go-wrapper", "run", "--output-dir=/crawls"] # ["app"]
