FROM golang:1.22-alpine3.18

WORKDIR /app

COPY . .

RUN apk add bash make musl-dev gcc

CMD [ "make run" ]