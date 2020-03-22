# ssh-tunnels
A specialized SSH tunneling server that get JWT token as password and validates against required Remote or Local forwarding with scopes

## Usage - shared key signing

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
     - LOG_LEVEL=debug
     - JWT_KEY=123
     - ENABLE_REMOTE_FORWARDING=true
     - ENABLE_LOCAL_FORWARDING=true
     - ENABLE_PTY=true
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


## Usage - RS512 pub/priv signing key

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
      - LOG_LEVEL=debug
      - JWT_
      - JWT_KEY_FILE=./test.key
      - ENABLE_REMOTE_FORWARDING=true
      - ENABLE_LOCAL_FORWARDING=true
      - ENABLE_PTY=true

secrets:
  rs-pub-key:
    file: ./test.pub
```

* Create file test.pub with public key contents

```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAssjOu4NJGni0f2vNqmTy
yVRXzHJmripzhIjGzaeI7pdBJImufTyV2+1cr+laiSMlXyNBvGWP/5/XAeM7Y0mX
GAnZkSrQvslQaAXR5E9aLEQjtDMumqVvXq6w+VZwVWmbYLYjX3WtTtztNqv3IsMk
0egpBzmJdixFZsjiFd3WlxvsZj/Zc9o+CaucDHRhOsJV3PqFp/aDKPKwKyZwMiXu
ZatzOJ8K4idj75PjnUEX+dp28ZB2boCdLwVES4uhqjC59YLe0UuZjNnIrordzcEk
G3l/bzZlZX54bYZ/1XcVWamDUPjKkXSBVjlweYrIOokuNoNLrNGFFzE18Gk30yp6
HwIDAQAB
-----END PUBLIC KEY-----
```

* Create a token at https://8gwifi.org/jwsgen.jsp with contents
```json
{
  "aid": "234243453453",
  "mid": "GTE3456",
  "exp": 1587529693,
  "rfw": "0.0.0.0:4343 0.0.0.0:4242",
  "lfw": "10.1.1.254:80 10.1.1.254:81 45.55.44.56:80",
  "pty": "true"
}
```

## ENVs

* JWT_ALGORITHM - JWT algorithm used for signing entered tokens. defaults to "HS512"
* JWT_KEY - key used by the signing algorith. required
* LOG_LEVEL - log level (error, warn, info, debug). defaults to info
* BIND_HOST - host to bind service to. defaults to 0.0.0.0 (all host interfaces will respond)
* BIND_PORT - ssh service port. defaults to 22
* ENABLE_REMOTE_FORWARDING - enable remote port forwarding. if not enabled, even if authorized on JWT token, it won't work. default. to false.
* ENABLE_LOCAL_FORWARDING - enable local port forwarding. if not enabled, event if authorized on JWT token, it won't work. defaults to false.
* ENABLE_PTY - enable pty terminal with a shell session on connect. if not enabled, even if authorized on JWT token, it won't work. defaults to true
* JWT_KEY_SECRET_NAME - Docker secret that will be used for loading key into ssh
