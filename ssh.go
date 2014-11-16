package main

import (
  "bytes"
  "code.google.com/p/go.crypto/ssh"
  "fmt"
  "io/ioutil"
  "os"
  "time"
)

// sshClient records necessary information to connect to a server.
type sshClient struct {
  Host   string
  Config *ssh.ClientConfig
  Client *ssh.Client
  Logger *logger
}

// execution records output of a SSH command and its duration.
type execution struct {
  Stdout  bytes.Buffer
  Stderr  bytes.Buffer
  Runtime time.Duration
}

// privateKey reads the personal private key, parse it and return a SSH signer.
func privateKey() ssh.Signer {
  pkFile := os.Getenv("HOME") + "/.ssh/id_rsa"

  key, err := ioutil.ReadFile(pkFile)
  if err != nil {
    fmt.Println("Failed to load private key.", err)
    os.Exit(1)
  }

  signer, err := ssh.ParsePrivateKey(key)
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }

  return signer
}

// newSSHClient returns a new sshClient instance with the given information.
func newSSHClient(host, user string) *sshClient {
  return &sshClient{
    Host: host,
    Config: &ssh.ClientConfig{
      User: user,
      Auth: []ssh.AuthMethod{
        ssh.PublicKeys(privateKey()),
      },
    },
    Logger: &logger{Prefix: host},
  }
}

// close closes the SSH client dialing if it is connected.
func (c *sshClient) close() {
  if c.Client != nil {
    c.Client.Close()
  }
}

// connect starts a connection with the server.
func (c *sshClient) connect() {
  c.Logger.warnf("Connecting to %s...", c.Host)

  var err error
  c.Client, err = ssh.Dial("tcp", c.Host+":22", c.Config)
  if err != nil {
    c.Logger.errorf("Failed to connect to %s. %s", c.Host, err.Error())
    os.Exit(1)
  }
}

// connectWhenNotConnected ensures that the client is connected with the server
// if it is not.
func (c *sshClient) connectWhenNotConnected() {
  if c.Client == nil {
    c.connect()
  }
}

// execute executes a shell command remotely given a number of commands.
func (c *sshClient) execute(cmd string) (*execution, error) {
  c.connectWhenNotConnected()

  started := time.Now()

  // Each ClientConn can support multiple interactive sessions,
  // represented by a Session.
  session, err := c.Client.NewSession()
  if err != nil {
    c.Logger.error("Failed to create session.")
    return nil, err
  }
  defer session.Close()

  // Once a Session is created, you can execute a single command on
  // the remote side using the Run method.
  var ex execution
  session.Stdout, session.Stderr = &ex.Stdout, &ex.Stderr

  c.Logger.warnf("Executing command `%s`.", cmd)
  if err := session.Run(cmd); err != nil {
    return &ex, err
  }

  ex.Runtime = time.Now().Sub(started)

  c.Logger.infof("Finished in %fs.", ex.Runtime.Seconds())

  return &ex, nil
}
