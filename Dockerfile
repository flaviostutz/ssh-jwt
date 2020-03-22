FROM golang:1.14.0-alpine3.11

ADD go.mod /ssh-jwt/
WORKDIR /ssh-jwt
RUN go mod download

ADD / /ssh-jwt
RUN go build -o /usr/bin/ssh-jwt

CMD [ "/startup.sh" ]

