FROM golang:1.14.0-alpine3.11

ADD go.mod /ssh-tunnels/
WORKDIR /ssh-tunnels
RUN go mod download

ADD / /ssh-tunnels
RUN go build -o /usr/bin/ssh-tunnels

CMD [ "/startup.sh" ]

