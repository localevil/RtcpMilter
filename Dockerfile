FROM golang:1.9 AS golang

ENV GIT_SSL_NO_VERIFY=1

RUN go get github.com/phalaaxx/milter

WORKDIR /go/src/app
COPY milter.go .

RUN go build

FROM registry.sigma.sbrf.ru/base/rhel/rhel7-minimal:latest

COPY --from=golang /go/src/app/app /usr/bin

EXPOSE 8080

ENTRYPOINT [ "/bin/sh" ]

CMD ["-c", "app -proto tcp -port :1000"]