package logger

import "github.com/sirupsen/logrus"

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		DisableHTMLEscape: true,
	})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetReportCaller(true)
}
