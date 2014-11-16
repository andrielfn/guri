package main

import (
  "fmt"
  "github.com/codegangsta/cli"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "os"
)

type config struct {
  Host         string
  User         string
  DeployTo     string "deploy_to"
  OS           string
  Arch         string
  Folders      []string
  Setup        []string
  BeforeDeploy []string "before_deploy"
  AfterDeploy  []string "after_deploy"
}

// deployConfig reads the deploy.yml file and returns a instance of config.
func deployConfig() *config {
  if _, err := os.Stat("deploy.yml"); os.IsNotExist(err) {
    fmt.Println("deploy.yml not found. Run `guri init` to create one.")
    os.Exit(1)
  }
  c := &config{}

  data, err := ioutil.ReadFile("deploy.yml")
  if err != nil {
    fmt.Println("Failed to read deploy.yml")
    os.Exit(1)
  }

  err = yaml.Unmarshal(data, c)
  if err != nil {
    fmt.Println("Failed to read deploy.yml")
    os.Exit(1)
  }

  return c
}

// newDeployFile creates a new config file if it does not exist.
// The file is created with some predefined data that should be replaced
// with real information.
func newDeployFile(c *cli.Context) {
  if _, err := os.Stat("deploy.yml"); os.IsNotExist(err) {
    // TODO: Move this config to a specific file.
    data := `# Server and build configuration.
host: example.com
user: root
deploy_to: /var/www/app

# Available OS are 'darwin', 'freebsd', 'linux' and 'windows'.
os: "linux"

# Available architectures are '386', 'amd64' and 'arm'.
arch: "amd64"

# The following folders will be attached to be deployed.
folders:
  - public
  - templates

# The following tasks will be executed with 'guri setup'.
# By default, the setup will create the 'deploy_to' folder.
setup:
  - "touch /var/www/app/app.log"

# The following tasks will be executed before the deploy.
before_deploy:
  - "service nginx stop"

# The following tasks will be executed after the deploy.
after_deploy:
  - "cd /var/www/app && nohup ./app > /var/www/app/app.log 2>&1 &"`

    ioutil.WriteFile("deploy.yml", []byte(data), 0644)
    fmt.Println("Created deploy.yml.\nEdit this file then run `guri setup`.")
  } else {
    fmt.Println("You already have a deploy.yml file.")
  }
}
