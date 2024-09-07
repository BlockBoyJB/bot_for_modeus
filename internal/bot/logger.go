package bot

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func setLogger(level, output string) *logrus.Logger {
	l := logrus.New()
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logLevel)
	}
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
	})
	if output == "stdout" {
		l.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Fatal(err)
		}
		l.SetOutput(file)
	}
	return l
}
