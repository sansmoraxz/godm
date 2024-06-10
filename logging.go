package godm

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	var logFile, err = os.Create(DefaultLogPath())
	if err != nil {
		panic(err)
	}
	log = &logrus.Logger{
		Out:          logFile,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: true,
	}
}

func DefaultLogPath() string {
	return os.TempDir() + string(os.PathSeparator) + "godm.log"
}
