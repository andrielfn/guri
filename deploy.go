package main

import (
  "bytes"
  "fmt"
  "github.com/codegangsta/cli"
  "io/ioutil"
  "os"
  "os/exec"
  "time"
)

type deploy struct {
  Config  *config
  Logger  *logger
  SSH     *sshClient
  Stderr  bytes.Buffer
  TempDir string
}

// build builds the binary following OS and ARCH predefined in the deploy config
// file. Note that you must have bootstrapped the OS/ARCH before the build.
//
// The binary file is placed in a OS temporary folder.
func (d *deploy) build() error {
  d.Logger.warnf("Building binary in %s.", d.TempDir)

  started := time.Now()

  goPath := fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH"))
  binOS := fmt.Sprintf("GOOS=%s", d.Config.OS)
  binArch := fmt.Sprintf("GOARCH=%s", d.Config.Arch)

  cmd := exec.Command("go", "build", "-o", d.TempDir+"/app")
  cmd.Env = []string{goPath, binOS, binArch}
  cmd.Stdout, cmd.Stderr = os.Stdout, &d.Stderr
  if err := cmd.Run(); err != nil {
    return err
  }

  d.Logger.infof("Finished in %fs.", time.Now().Sub(started).Seconds())

  return nil
}

// folders attach any folder predefined in deploy config file to be deployed
// together with the binary.
//
// The folders are copied to a OS temporary folder, same folder the binary was
// placed.
func (d *deploy) folders() error {
  for _, folder := range d.Config.Folders {
    d.Logger.warnf("Copying folder `%s`.", folder)

    started := time.Now()

    cmd := exec.Command("cp", "-R", folder, d.TempDir)
    cmd.Stdout, cmd.Stderr = os.Stdout, &d.Stderr
    if err := cmd.Run(); err != nil {
      return err
    }

    d.Logger.infof("Finished in %fs.", time.Now().Sub(started).Seconds())
  }

  return nil
}

// compress compress the temporary folder with the binary and attached folders
// to a tarball file called deploy.tar.gz.
//
// The tar file is placed in the same directory of the application.
func (d *deploy) compress() error {
  d.Logger.warnf("Compressing `%s` to `%s`.", d.TempDir, "deploy.tar.gz")

  started := time.Now()

  cmd := exec.Command("tar", "-cvzf", "deploy.tar.gz", "-C", d.TempDir, ".")
  if err := cmd.Run(); err != nil {
    return err
  }

  d.Logger.infof("Finished in %fs.", time.Now().Sub(started).Seconds())

  return nil
}

// transfer sends the compressed tar file to the remote server with SCP.
//
// The file is sent to the `deploy_to` folder predefined in the config file.
func (d *deploy) transfer() error {
  d.Logger.warnf("Transfering package to remote path: %s.", d.Config.DeployTo)

  started := time.Now()
  remoteAddr := fmt.Sprintf("%s@%s:%s", d.Config.User, d.Config.Host, d.Config.DeployTo)

  cmd := exec.Command("scp", "deploy.tar.gz", remoteAddr)
  cmd.Stdout, cmd.Stderr = os.Stdout, &d.Stderr
  if err := cmd.Run(); err != nil {
    return err
  }

  d.Logger.infof("Finished in %fs.", time.Now().Sub(started).Seconds())

  return nil
}

// extract extracts the tar file in the remote server.
func (d *deploy) extract() error {
  d.Logger.warn("Extracting the package in the remote server.")

  cmd := fmt.Sprintf("tar zxf %s/deploy.tar.gz -C %s", d.Config.DeployTo, d.Config.DeployTo)

  ex, err := d.SSH.execute(cmd)
  if err != nil {
    d.Logger.errorf("Failed to execute `%s`.", cmd)
    d.Logger.error(ex.Stderr.String())
    return err
  }

  return nil
}

// beforeDeploy runs every task predefined in the deploy config file BEFORE the
// deploy.
//
// The shell command is runned the same way is was described in the config file.
func (d *deploy) beforeDeploy() error {
  for _, cmd := range d.Config.BeforeDeploy {
    ex, err := d.SSH.execute(cmd)
    if err != nil {
      d.Logger.errorf("Failed to execute `%s`.", cmd)
      d.Logger.error(ex.Stderr.String())
      return err
    }
  }

  return nil
}

// afterDeploy runs every task predefined in the deploy config file AFTER the
// deploy.
//
// The shell command is runned the same way is was described in the config file.
func (d *deploy) afterDeploy() error {
  for _, cmd := range d.Config.AfterDeploy {
    ex, err := d.SSH.execute(cmd)
    if err != nil {
      d.Logger.errorf("Failed to execute `%s`.", cmd)
      d.Logger.error(ex.Stderr.String())
      return err
    }
  }

  return nil
}

// cleanup removes the temporary folder created and the tar file with the deploy
// package.
func (d *deploy) cleanup() error {
  files := []string{"deploy.tar.gz", d.TempDir}

  for _, f := range files {
    if _, err := os.Stat(f); os.IsNotExist(err) {
      continue
    }

    d.Logger.warnf("Removing `%s` locally.", f)
    started := time.Now()

    cmd := exec.Command("rm", "-Rf", f)
    if err := cmd.Run(); err != nil {
      return err
    }

    d.Logger.infof("Finished in %fs.", time.Now().Sub(started).Seconds())
  }

  return nil
}

// runDeploy delegates every deploy task in the correctly order.
//
// If any task catch an error, the deploy will be interrupted and the cleanup
// method will be called.
func runDeploy(c *cli.Context) {
  started := time.Now()

  config := deployConfig()
  tempDir, _ := ioutil.TempDir("", "guri-")
  ssh := newSSHClient(config.Host, config.User)

  // New deploy instance.
  d := &deploy{
    Config:  config,
    SSH:     ssh,
    Logger:  &logger{Prefix: config.Host},
    TempDir: tempDir,
  }

  // Build.
  if err := d.build(); err != nil {
    d.Logger.error(d.Stderr.String())
    d.cleanup()
    os.Exit(1)
  }

  // Folders.
  if err := d.folders(); err != nil {
    d.Logger.error(d.Stderr.String())
    d.cleanup()
    os.Exit(1)
  }

  // Compress.
  if err := d.compress(); err != nil {
    d.Logger.error(d.Stderr.String())
    d.cleanup()
    os.Exit(1)
  }

  // Before deploy.
  if err := d.beforeDeploy(); err != nil {
    d.Logger.error(err)
    d.cleanup()
    os.Exit(1)
  }

  // Transfer.
  if err := d.transfer(); err != nil {
    d.Logger.error(d.Stderr.String())
    d.cleanup()
    os.Exit(1)
  }

  // Extract.
  if err := d.extract(); err != nil {
    d.Logger.error(err)
    d.cleanup()
    os.Exit(1)
  }

  // After deploy.
  if err := d.afterDeploy(); err != nil {
    d.Logger.error(err)
    d.cleanup()
    os.Exit(1)
  }

  d.cleanup()
  d.SSH.close()

  d.Logger.infof("Done in %f.", time.Now().Sub(started).Seconds())
}
