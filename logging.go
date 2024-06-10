package godm

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

type myFormatter struct {
    logrus.TextFormatter
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    return []byte(fmt.Sprintf("[%s] - %s\t%s\n",
		entry.Time.Format(time.RFC3339), strings.ToUpper(entry.Level.String()), entry.Message)), nil
}


func init() {
	var logFile, err = os.Create(DefaultLogPath())
	if err != nil {
		panic(err)
	}
	fmtr := new(myFormatter)
	log = &logrus.Logger{
		Out:          logFile,
		Formatter:    fmtr,
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: true,
	}
}

func DefaultLogPath() string {
	return os.TempDir() + string(os.PathSeparator) + "godm.log"
}
