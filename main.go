package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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

type options struct {
	sshBindHost                string
	sshPort                    int
	enableLocalPortForwarding  bool
	enableRemotePortForwarding bool
	enablePty                  bool
	jwtSignatureAlgorithm      jwa.SignatureAlgorithm
	jwtKey                     string
	jwtKeyFile                 string
}

func main() {

	logLevel := flag.String("log-level", "debug", "debug, info, warning, error")
	sshBindHost0 := flag.String("bind-host", "0.0.0.0", "Bind host for SSH service. Defaults to 0.0.0.0")
	sshPort0 := flag.Int("port", 22, "SSH server port to listen on. Defaults to 22")
	enableRemotePortForwarding0 := flag.Bool("enable-remote-forwarding", false, "Enable remote port forwarding bind. Defaults to false")
	enableLocalPortForwarding0 := flag.Bool("enable-local-forwarding", false, "Enable local port forwarding. Defaults to false")
	enablePty0 := flag.Bool("enable-pty", false, "Enable PTY")
	jwtSignatureAlgorithm0 := flag.String("jwt-algorithm", "HS512", "JWT signature algorithm. Defaults to HS512")
	jwtKey0 := flag.String("jwt-key", "", "JWT key contents. Required if jwt-key-file is not defined")
	jwtKeyFile0 := flag.String("jwt-key-file", "", "JWT key file. Required")
	flag.Parse()

	switch *logLevel {
	case "trace":
		logrus.SetLevel(logrus.TraceLevel)
		break
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	sa := jwa.SignatureAlgorithm(*jwtSignatureAlgorithm0)
	err := sa.Accept(*jwtSignatureAlgorithm0)
	if err != nil {
		logrus.Errorf("JWT signing algorithm is not supported. err=%s", err)
		panic("")
	}

	opt := options{
		sshPort:                    *sshPort0,
		enableRemotePortForwarding: *enableRemotePortForwarding0,
		enableLocalPortForwarding:  *enableLocalPortForwarding0,
		enablePty:                  *enablePty0,
		sshBindHost:                *sshBindHost0,
		jwtSignatureAlgorithm:      sa,
		jwtKeyFile:                 *jwtKeyFile0,
		jwtKey:                     *jwtKey0,
	}

	//load key contents
	var jwtKeyContents []byte
	if opt.jwtKey == "" {
		if opt.jwtKeyFile != "" {
			jwtKeyContents, err = ioutil.ReadFile(opt.jwtKeyFile)
			if err != nil {
				logrus.Errorf("Couldn't read key file contents. err=%s", err)
				return
			}
			logrus.Debugf("JWT key loaded from file '%s'", opt.jwtKeyFile)
		} else {
			logrus.Errorf("Either --jwt-key-file of --jwt-key is required")
			return
		}
	} else {
		jwtKeyContents = []byte(opt.jwtKey)
		logrus.Debugf("JWT key loaded from 'jwt-key' arg")
	}

	jwtKey, err := parsePKIXPublicKeyFromPEM(jwtKeyContents)
	if err != nil {
		logrus.Debugf("Couldn't parse public key from PEM. err=%s", err)
		jwtKey = jwtKeyContents
	}

	logrus.Infof("Starting ssh server on port %s:%d...", opt.sshBindHost, opt.sshPort)
	forwardHandler := &ssh.ForwardedTCPHandler{}

	if opt.enableLocalPortForwarding {
		logrus.Infof("Local port forwarding is enabled")
	} else {
		logrus.Infof("Local port forwarding is disabled")
	}

	if opt.enableRemotePortForwarding {
		logrus.Infof("Remote port forwarding is enabled")
	} else {
		logrus.Infof("Remote port forwarding is disabled")
	}

	if opt.enablePty {
		logrus.Infof("PTY is enabled")
	} else {
		logrus.Infof("PTY is disabled")
	}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			if !opt.enableLocalPortForwarding {
				logrus.Debugf("Local port forwarding is disabled")
				return false
			}
			claim0 := ctx.Value("lfw")
			if claim0 != nil {
				claim := claim0.(string)
				if claim == "" {
					logrus.Infof("No local forward claims found in token")
					return false
				}

				accept := matchClaim(claim, dhost, dport)

				if !accept {
					logrus.Infof("Denying local port forward %s:%d (direct-tcpip)", dhost, dport)
					return false
				}
				logrus.Debugf("Allowing local port forward %s:%d (direct-tcpip)", dhost, dport)

				return true
			}
			return false
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, bindHost string, port uint32) bool {
			if !opt.enableRemotePortForwarding {
				logrus.Debugf("Remote port forwarding is disabled")
				return false
			}
			claim0 := ctx.Value("rfw")
			if claim0 != nil {
				claim := claim0.(string)
				if claim == "" {
					logrus.Infof("No remote forward claims found in token")
					return false
				}

				accept := matchClaim(claim, bindHost, port)

				if !accept {
					logrus.Infof("Denying remote port forwarding %s:%d (tcpip-forward)", bindHost, port)
					return false
				}
				logrus.Debugf("Allowing remote port forwarding %s:%d (tcpip-forward)", bindHost, port)

				return true
			}
			return false
		}),
		PtyCallback: ssh.PtyCallback(func(ctx ssh.Context, pty ssh.Pty) bool {
			if !opt.enablePty {
				return false
			}
			pty0 := ctx.Value("pty")
			if pty0 != nil {
				pty1 := pty0.(string)
				if pty1 == "true" {
					logrus.Debugf("Denying PTY")
					return true
				}
			}
			logrus.Debugf("Allowing PTY")
			return false
		}),
		PasswordHandler: ssh.PasswordHandler(func(ctx ssh.Context, password string) bool {
			//SAMPLE: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMiLCJsZnciOiIyMDEuMjEuNDMuNDU6ODA4MCJ9.iaUGlrO-3HWdE-8irizqMfHLYV0Ctiu3N3qdEdirwJk
			//SAMPLE2: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhaWQiOiIyMzQyNDM0NTM0NTMiLCJtaWQiOiJHVEUzNDU2IiwiZXhwIjoxNTg3NTI5NjkzLCJyZnciOiIwLjAuMC4wOjQzNDMiLCJsZnciOiIxMC4xLjEuMjU0OjgwIn0.ynmGKtRJyr5KowmD34m3A4OBnMdcmj9GCC0Vt3oyZHc
			tokenString := password
			token, err := jwt.Parse(bytes.NewReader([]byte(tokenString)), jwt.WithVerify(opt.jwtSignatureAlgorithm, jwtKey))
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
			rfw, ok := token.Get("rfw")
			if ok {
				ctx.SetValue("rfw", rfw)
			} else {
				ctx.SetValue("rfw", "")
			}

			//get local forward permission
			lfw, ok := token.Get("lfw")
			if ok {
				ctx.SetValue("lfw", lfw)
			} else {
				ctx.SetValue("lfw", "")
			}

			//get PTY permission
			pty, ok := token.Get("pty")
			if ok {
				ctx.SetValue("pty", pty)
			} else {
				ctx.SetValue("pty", "false")
			}

			logrus.Debugf("Valid token received. remoteForward=%s. localForward=%s pty=%s", rfw, lfw, pty)

			return true
		}),
		Addr: fmt.Sprintf("%s:%d", opt.sshBindHost, opt.sshPort),
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
		// IdleTimeout: 1 * time.Second,
		// MaxTimeout:  3 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		logrus.Errorf("Startup error %s", err)
	}
}

func matchClaim(claim string, host string, port uint32) bool {
	claim = strings.ReplaceAll(claim, "localhost", "127.0.0.1")
	host = strings.ReplaceAll(host, "localhost", "127.0.0.1")
	cps := strings.Split(claim, " ")
	accept := false
	for _, cp := range cps {
		logrus.Tracef("claim: %v -- host: %v -- port: %v", claim, host, port)
		if cp == fmt.Sprintf("%s:%d", host, port) {
			logrus.Trace("accepted")
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

func parsePKIXPublicKeyFromPEM(pubPEM []byte) (interface{}, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}
