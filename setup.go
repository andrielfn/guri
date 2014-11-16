package main

import (
  "fmt"
  "github.com/codegangsta/cli"
  "time"
)

// setup creates the `deploy_to` folder predefined in the deploy file in remote
// server.
//
// If the `deploy_to` is configured to something like `/var/www/my-app`, the
// setup will remotely run `mkdir -p /var/www/my-app`. It also assumes that the
// `deploy_to` folder is fully writable/readable for the user we are going to
// SSH with.
//
// Also, if any setup task was predefined in the config file, it will run here
// the same way it was described there.
func setupServer(c *cli.Context) {
  started := time.Now()

  config := deployConfig()
  logger := &logger{Prefix: config.Host}
  ssh := newSSHClient(config.Host, config.User)

  ssh.execute(fmt.Sprintf("mkdir -p %s", config.DeployTo))

  for _, t := range config.Setup {
    ssh.execute(t)
  }

  ssh.close()

  logger.infof("Done in %f.", time.Now().Sub(started).Seconds())
}
