FROM golang:1.14.0-alpine3.11

ENV LOG_LEVEL 'info'
ENV BIND_HOST '0.0.0.0'
ENV BIND_PORT '22'
ENV ENABLE_REMOTE_FORWARDING 'false'
ENV ENABLE_LOCAL_FORWARDING 'false'
ENV ENABLE_PTY 'false'
ENV JWT_ALGORITHM 'HS512'
ENV JWT_KEY ''
ENV JWT_KEY_SECRET_NAME ''

WORKDIR /ssh-jwt

ADD go.mod /ssh-jwt/
RUN go mod download

ADD / /ssh-jwt
RUN go build -o /usr/bin/ssh-jwt

ADD startup.sh /

CMD [ "/startup.sh" ]

