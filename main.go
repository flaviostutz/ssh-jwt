package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
)

func main() {

	log.Println("Starting ssh server on port 2222...")

	forwardHandler := &ssh.ForwardedTCPHandler{}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			claim0 := ctx.Value("lfw")
			claim := claim0.(string)
			if claim == "" {
				logrus.Infof("No local forward claims found in JWT")
				return false
			}

			cps := strings.Split(claim, " ")
			accept := false
			for _, cp := range cps {
				c1cs := strings.Split(cp, ":")
				if c1cs[0] == fmt.Sprintf("%s:%d", dhost, dport) {
					accept = true
					break
				}
			}
			if !accept {
				logrus.Debugf("Forward %s:%s is NOT authorized", dhost, dport)
				return false
			}
			logrus.Debugf("Forward %s:%s authorized", dhost, dport)

			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			logrus.Debugf("Attempt to bind %s:%s", host, port)
			return true
		}),
		PasswordHandler: ssh.PasswordHandler(func(ctx ssh.Context, password string) bool {
			//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMiLCJsZnciOiIyMDEuMjEuNDMuNDU6ODA4MCJ9.iaUGlrO-3HWdE-8irizqMfHLYV0Ctiu3N3qdEdirwJk
			tokenString := password
			token, err := jwt.Parse(bytes.NewReader([]byte(tokenString)), jwt.WithVerify(jwa.HS256, []byte("123")))
			if err != nil {
				logrus.Infof("Failed to parse JWT token. err=%s", err)
				return false
			}
			err = token.Verify()
			if err != nil {
				logrus.Infof("Invalid token received. err=%s", err)
				return false
			}

			//get remote forward permission
			one := false
			rfw, ok := token.Get("rfw")
			if ok {
				one = true
				ctx.SetValue("rfw", rfw)
			} else {
				ctx.SetValue("rfw", "")
			}

			//get local forward permission
			lfw, ok := token.Get("lfw")
			if ok {
				one = true
				ctx.SetValue("lfw", lfw)
			} else {
				ctx.SetValue("lfw", "")
			}

			if !one {
				logrus.Infof("Invalid token received. It must have either 'lfw' (local forward claim) or 'rfw' (remote forward claim). Ex.: lfw=201.234.32.11:3455 rfw=0.0.0.0:34938.")
				return false
			}

			logrus.Debugf("Valid token received. remoteForward=%s. localForward=%s", rfw, lfw)

			return true
		}),
		Addr: ":2222",
		Handler: ssh.Handler(func(s ssh.Session) {
			io.WriteString(s, "Tunnels prepared. Waiting for connections...\n")
			select {}
		}),
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		// IdleTimeout: 30 * time.Second,
		// MaxTimeout:  120 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
