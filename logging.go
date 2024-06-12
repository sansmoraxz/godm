package godm

import (
	"fmt"
	"io"
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
	var logFile, err = os.OpenFile(DefaultLogPath(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

func ViewLog() error {
	file, err := os.Open(DefaultLogPath())
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		var buffer = make([]byte, 1024)
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Print(string(buffer[:n]))
	}
	return nil
}


func ClearLog() error {
	// empty the log file content
	var logFile, err = os.OpenFile(DefaultLogPath(), os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer logFile.Close()
	return nil
}
