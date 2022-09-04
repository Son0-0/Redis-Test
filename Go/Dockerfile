FROM golang

RUN mkdir /app

ADD . /app

WORKDIR /app

RUN go build -o main .

EXPOSE 9090

CMD [ "/app/main" ]