package slackchatops

import (
	logrus "github.com/sirupsen/logrus"
)

//const sentryDSN = "https://xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

func NewLogger(app string) *logrus.Entry {

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		//ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		//DisableColors:   true,
	}
	// &logrus.JSONFormatter{}

	// hook, err := logrus_sentry.NewSentryHook(sentryDSN, []logrus.Level{
	// 	logrus.PanicLevel,
	// 	logrus.FatalLevel,
	// 	logrus.ErrorLevel,
	// })
	// if err == nil {
	// 	log.Hooks.Add(hook)
	// }
	logger := log.WithFields(logrus.Fields{"app": app})
	return logger

}
