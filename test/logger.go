package test

import "fmt"

type logger struct {
	collected  string
	lastLog    string
	logChannel chan string
}

func NewLogger() *logger {
	return &logger{
		collected:  "",
		lastLog:    "",
		logChannel: make(chan string, 100),
	}
}

func (d *logger) Collected() string {
	return d.collected
}

func (d *logger) Prune() {
	d.collected = ""
	d.lastLog = ""
}

func (d *logger) LastLog() string {
	return d.lastLog
}

func (d *logger) LogChannel() <-chan string {
	return d.logChannel
}

func (d *logger) log(level string, f string, v ...interface{}) {
	if f[len(f)-1] != '\n' {
		f += "\n"
	}
	log := fmt.Sprintf(level+": "+f, v...)
	d.collected += log
	d.lastLog = log
	d.logChannel <- log
}

func (d *logger) Tracef(f string, v ...interface{}) {
	d.log("TRACE", f, v...)
}

func (d *logger) Debugf(f string, v ...interface{}) {
	d.log("DEBUG", f, v...)
}

func (d *logger) Infof(f string, v ...interface{}) {
	d.log("INFO", f, v...)
}

func (d *logger) Warnf(f string, v ...interface{}) {
	d.log("WARN", f, v...)
}

func (d *logger) Errorf(f string, v ...interface{}) {
	d.log("ERROR", f, v...)
}

func (d *logger) Fatalf(f string, v ...interface{}) {
	d.log("FATAL", f, v...)
}
