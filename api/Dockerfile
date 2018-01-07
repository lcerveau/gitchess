FROM golang

ADD ./src /go/src/com.mmyneta/golang/gitchess

RUN go install com.mmyneta/golang/gitchess

ENTRYPOINT /go/bin/gitchess

EXPOSE 8080

