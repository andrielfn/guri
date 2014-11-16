package main

import (
  "testing"
)

func TestLog(t *testing.T) {
  logger := &logger{Prefix: "hostname"}

  logger.warn("asd")

  // if debug != "debug log" {
  //   t.Error("lalalal")
  // }
}
