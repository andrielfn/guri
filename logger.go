package main

import (
  "fmt"
  "time"
)

var logColors = map[int]int{
  DEBUG: 102,
  INFO:  28,
  WARN:  214,
  ERROR: 196,
}

const TIME_FORMAT = "2006/01/02 15:04:05"

func colorize(c int, s string) (r string) {
  return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c, s)
}

const (
  DEBUG = iota
  INFO
  WARN
  ERROR
)

var logPrefixes = map[int]string{
  DEBUG: "DEBUG",
  INFO:  "INFO",
  WARN:  "WARN",
  ERROR: "ERROR",
}

type logger struct {
  LogLevel int
  Prefix   string
}

func (l *logger) debugf(format string, n ...interface{}) {
  l.logf(DEBUG, format, n...)
}

func (l *logger) infof(format string, n ...interface{}) {
  l.logf(INFO, format, n...)
}

func (l *logger) warnf(format string, n ...interface{}) {
  l.logf(WARN, format, n...)
}

func (l *logger) errorf(format string, n ...interface{}) {
  l.logf(ERROR, format, n...)
}

func (l *logger) logf(level int, s string, n ...interface{}) {
  if level >= l.LogLevel {
    fmt.Println(l.logPrefix(level), fmt.Sprintf(s, n...))
  }
}

func (l *logger) debug(n ...interface{}) {
  l.log(DEBUG, n...)
}

func (l *logger) info(n ...interface{}) {
  l.log(INFO, n...)
}

func (l *logger) warn(n ...interface{}) {
  l.log(WARN, n...)
}

func (l *logger) error(n ...interface{}) {
  l.log(ERROR, n...)
}

func (l *logger) logPrefix(i int) (s string) {
  s = time.Now().Format(TIME_FORMAT)

  if l.Prefix != "" {
    s = s + " [" + l.Prefix + "]"
  }

  s = s + " " + l.logLevelPrefix(i)
  return
}

func (l *logger) logLevelPrefix(level int) (s string) {
  color := logColors[level]
  prefix := logPrefixes[level]
  return colorize(color, prefix)
}

func (l *logger) log(level int, n ...interface{}) {
  if level >= l.LogLevel {
    all := append([]interface{}{l.logPrefix(level)}, n...)
    fmt.Println(all...)
  }
}
