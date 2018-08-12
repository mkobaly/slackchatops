package slackchatops

import (
	logrus "github.com/sirupsen/logrus"
)

func NewLogger(app string) *logrus.Entry {

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		//ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		//DisableColors:   true,
	}
	logger := log.WithFields(logrus.Fields{"app": app})
	return logger
}
