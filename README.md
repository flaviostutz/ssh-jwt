# ssh-tunnels
A specialized SSH tunneling server that get JWT token as password and validates against required Remote or Local forwarding with scopes

## Usage

* create docker-compose.yml:

```yml
version: '3.7'
services:
  ssh-tunnels:
    image: flaviostutz/ssh-jwt
    ports:
      - "2222:22"
    restart: always
    environment:
     --LOG_LEVEL=debug
     --ENABLE_REMOTE_FORWARDING=true
     --ENABLE_LOCAL_FORWARDING=true
     --ENABLE_PTY=true
```

* run docker-compose up

* In a terminal, run

```bash
ssh root@localhost -p 2222
```

* When asked for password, paste

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMgMC4wLjAuMDo0MjQyIiwibGZ3IjoiMTAuMS4xLjI1NDo4MCAxMC4xLjEuMjU0OjgxIDQ1LjU1LjQ0LjU2OjgwIiwicHR5IjoidHJ1ZSJ9.wVQ46URtFFntfwfxJKGNgXoDLvFFzvV-HQGOsM0-SHg
```

* If all is OK, you will be connected to a shell session.

* Inspect token contents to view its contents by opening https://jwt.io/ and pasting the above contents to view token structure

* Modify it to set "pty" to "false" and try to connect again

* If the token is invalid or it doesn't have claim "pty", you connection will be refused.


## JWT token Claims

* rfw - a list of space separated "[bindhost]:[port]" indicating authorized remote port forward binds. ex.: "0.0.0.0:4444" will accept remote port forwarding to the other side

* lfw - a list of space separated "[desthost]:[port]" indicating authorized local port forwards destinations. ex.: "201.22.123.43:80" will accept local port forwardings to 201.22.123.43 through the ssh tunnel

* pty - permit interactive terminal sessions in shell if "true"



## ENVs

* LOG_LEVEL - log level (error, warn, info, debug). defaults to info
* BIND_HOST - host to bind service to. defaults to 0.0.0.0 (all host interfaces will respond)
* BIND_PORT - ssh service port. defaults to 22
* ENABLE_REMOTE_FORWARDING - enable remote port forwarding. if not enabled, even if authorized on JWT token, it won't work. default. to false.
* ENABLE_LOCAL_FORWARDING - enable local port forwarding. if not enabled, event if authorized on JWT token, it won't work. defaults to false.
* ENABLE_PTY - enable pty terminal with a shell session on connect. if not enabled, even if authorized on JWT token, it won't work. defaults to true


