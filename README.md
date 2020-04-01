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

* Open http://jwt.io

* Create a JWT key with the following contents

header
```json
{
  "alg": "HS512",
  "typ": "JWT"
}
```

payload
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

signature
use password "123"

* Copy the encoded/signed JWT contents to clipboard

* In a terminal, run

```bash
ssh root@localhost -p 2222
```

* When asked for password, paste Enconded JWT contents

* If all is OK, you will be connected to a shell session.

* Modify JWT claim "pty" to "false" and try to connect again

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
    secrets:
      - rs-pub-key
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

* Open https://8gwifi.org/jwsgen.jsp

* Create a JWT key with the following contents

JWS Algo: RS512

Payload
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

Click on "Generate JWS Keys"

* Create file test_rsa.pub with public key contents

* Create file test_rsa.key with private key contents

* Copy the contents of the JWT key from the "Serialize" field from the site

* Run docker-compose up

* On another terminal, run
  * ```ssh root@localhost -p 2222 -L 0.0.0.0:1212:10.1.1.254:80```

* On a third terminal run
  * ```curl localhost:1212```

* If any web server is running on 10.1.1.254:80 it will get its contents


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
