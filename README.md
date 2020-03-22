# ssh-tunnels
A specialized SSH tunneling server that get JWT token as password and validates against required Remote or Local forwarding with scopes

## Usage

* create docker-compose.yml:

```yml
version: '3.7'
services:
  ssh-tunnels:
    image: flaviostutz/ssh-tunnels
    ports:
      - "2222:2222"
    restart: always
    environment:
      - JWT_PUB_KEY=123
```

