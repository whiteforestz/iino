FROM library/golang:1.17

WORKDIR ${GOPATH}/src/github.com/whiteforestz/iino/
COPY . ${GOPATH}/src/github.com/whiteforestz/iino/

RUN go build -o ${GOPATH}/bin/service ./cmd/service

CMD ${GOPATH}/bin/service
