FROM golang:1.19-alpine

RUN apk add alpine-sdk

RUN mkdir /app

COPY . /app/
WORKDIR /app
RUN go build .

RUN chmod +x /app/abs-specific

ENTRYPOINT [ "/app/abs-specific" ]
