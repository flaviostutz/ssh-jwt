package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	logrus.Infof("Starting ssh server on port 2222...")

	forwardHandler := &ssh.ForwardedTCPHandler{}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			claim0 := ctx.Value("lfw")
			claim := claim0.(string)
			if claim == "" {
				logrus.Infof("No local forward claims found in token")
				return false
			}

			accept := matchClaim(claim, dhost, dport)

			if !accept {
				logrus.Infof("Forward %s:%d is NOT authorized (direct-tcpip)", dhost, dport)
				return false
			}
			logrus.Debugf("Forward %s:%d is authorized (direct-tcpip)", dhost, dport)

			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, bindHost string, port uint32) bool {
			claim0 := ctx.Value("lfw")
			claim := claim0.(string)
			if claim == "" {
				logrus.Infof("No remote forward claims found in token")
				return false
			}

			accept := matchClaim(claim, bindHost, port)

			if !accept {
				logrus.Infof("Remote bind %s:%d is NOT authorized (tcpip-forward)", bindHost, port)
				return false
			}
			logrus.Debugf("Remote bind %s:%d is authorized (tcpip-forward)", bindHost, port)

			return true
		}),
		PasswordHandler: ssh.PasswordHandler(func(ctx ssh.Context, password string) bool {
			//SAMPLE: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMiLCJsZnciOiIyMDEuMjEuNDMuNDU6ODA4MCJ9.iaUGlrO-3HWdE-8irizqMfHLYV0Ctiu3N3qdEdirwJk
			//SAMPLE2: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMiLCJsZnciOiIxMC4xLjEuMjU0OjgwIn0.ynmGKtRJyr5KowmD34m3A4OBnMdcmj9GCC0Vt3oyZHc
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
		PtyCallback: ssh.PtyCallback(func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		}),
		Handler: ssh.Handler(func(s ssh.Session) {
			io.WriteString(s, "Waiting for connections...\n")

			cmd := exec.Command("sh")
			ptyReq, winCh, isPty := s.Pty()
			if isPty {
				cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
				f, err := pty.Start(cmd)
				if err != nil {
					panic(err)
				}
				go func() {
					for win := range winCh {
						setWinsize(f, win.Width, win.Height)
					}
				}()
				go func() {
					io.Copy(f, s) // stdin
				}()
				io.Copy(s, f) // stdout
				cmd.Wait()
			} else {
				io.WriteString(s, "No PTY requested.\n")
				s.Exit(1)
			}
		}),
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip": ssh.DirectTCPIPHandler,
			"session":      ssh.DefaultSessionHandler,
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		// IdleTimeout: 30 * time.Second,
		// MaxTimeout:  120 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		logrus.Errorf("Startup error %s", err)
	}
}

func matchClaim(claim string, host string, port uint32) bool {
	cps := strings.Split(claim, " ")
	accept := false
	for _, cp := range cps {
		if cp == fmt.Sprintf("%s:%d", host, port) {
			accept = true
			break
		}
	}
	return accept
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}
