# ssh-jwt
A SSH server that authorizes PTY, remote or local port forwarding based on JWT token entered as password.

View ENVs and Usage for more information on how to launch this.

## Usage

### shared key signing

* create docker-compose.yml:

```yml
version: '3.7'
services:
  ssh-tunnels:
    image: flaviostutz/ssh-jwt
    ports:
      - "2222:22"
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

### RS512 pub/priv signing key

* create docker-compose.yml:

```yml
version: '3.7'
services:
  ssh-tunnels:
    image: flaviostutz/ssh-jwt
    ports:
      - "2222:22"
    environment:
      - JWT_ALGORITHM=RS512
      - JWT_KEY_SECRET_NAME=rs-pub-key
      - ENABLE_LOCAL_FORWARDING=true
      - ENABLE_REMOTE_FORWARDING=true
      - ENABLE_PTY=true
      - LOG_LEVEL=debug

secrets:
  rs-pub-key:
    file: ./test_rsa.pub
```

* Create file test_rsa.pub with public key contents

```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArha48sD6KZfBbRQBeMpm
io4VOFu7hCCdJ0ICl845En2/IXvItgfmVXJd+h0aoZV9PBzW9l65ROfvEMmLrtla
DSCXLDnQwkc0NLGW0s4EdLR5wnUOgAuc4/Pp/pOEJATsc/JZxXPUbU2delMi9uYB
Jfgo/jeh0HGnDVi9dboZdjfRNndRQDJkEdBEVM9jHmTSZROsmgSem1tlrNT5Jw0u
SaSXxRYb3qo8A7044Ck+P436iprfNm2AgOLHcynjtZSKoLerAACh+7ZdcWPYLCB9
4ynKBAbhCme0Rc0rpexF+ChjaDLmWJumEFkgRKPohGm7jUTfdH5uHx27AKMMBUjh
yQIDAQAB
-----END PUBLIC KEY-----
```

* Run docker-compose up

* On another terminal, run
  * ```ssh root@localhost -p 2222 -L 0.0.0.0:1212:10.1.1.254:80```

* On a third terminal run
  * ```curl localhost:1212```

* If any web server is running on 10.1.1.254:80 it will get its contents

* Explore other tokens by playing with https://8gwifi.org/jwsgen.jsp. Token contents
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

## JWT token Claims

* rfw - a list of space separated "[bindhost]:[port]" indicating authorized remote port forward binds. ex.: "0.0.0.0:4444" will accept remote port forwarding to the other side

* lfw - a list of space separated "[desthost]:[port]" indicating authorized local port forwards destinations. ex.: "201.22.123.43:80" will accept local port forwardings to 201.22.123.43 through the ssh tunnel

* pty - permit interactive terminal sessions in shell if "true"


## ENVs

* JWT_ALGORITHM - JWT algorithm used for signing entered tokens. Maybe one of ES256, ES384, ES512, HS256, HS384, HS512, PS256, PS384, PS384, PS512, RS256, RS384, RS512. defaults to "HS512".
* JWT_KEY - key used by the signing algorith. required
* LOG_LEVEL - log level (error, warn, info, debug). defaults to info
* BIND_HOST - host to bind service to. defaults to 0.0.0.0 (all host interfaces will respond)
* BIND_PORT - ssh service port. defaults to 22
* ENABLE_REMOTE_FORWARDING - enable remote port forwarding. if not enabled, even if authorized on JWT token, it won't work. default. to false.
* ENABLE_LOCAL_FORWARDING - enable local port forwarding. if not enabled, event if authorized on JWT token, it won't work. defaults to false.
* ENABLE_PTY - enable pty terminal with a shell session on connect. if not enabled, even if authorized on JWT token, it won't work. defaults to true
* JWT_KEY_SECRET_NAME - Docker secret that will be used for loading key into ssh
