package main

import (
  "github.com/codegangsta/cli"
  "os"
)

func main() {
  app := cli.NewApp()

  app.Name = "guri"
  app.Usage = "A command-line interface to improve Go deployments."
  app.Version = "0.0.1"
  app.Author = "Andriel Nuernberg"
  app.Email = "andrielfn@gmail.com"

  app.Commands = []cli.Command{
    {
      Name:   "init",
      Usage:  "Create a new deploy config file in the current directory.",
      Action: newDeployFile,
    },
    {
      Name:   "setup",
      Usage:  "Create the `deploy_to` folder and running the predefined tasks defined in the deploy config file.",
      Action: setupServer,
    },
    {
      Name:   "deploy",
      Usage:  "Build the app, attach any predefined folder, compress the files and send it to server.",
      Action: runDeploy,
    },
  }

  app.Run(os.Args)
}
