
package openflow

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	)

type klogFormatter struct{}

// Format formats logrus log in compliance with k8s log.
func (f *klogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	filePath := entry.Caller.File
	filePathArray := strings.Split(filePath, "/")
	fileName := filePathArray[len(filePathArray)-1]
	pidString := strconv.Itoa(os.Getpid())

	// logrus has seven logging levels: Trace, Debug, Info, Warning, Error, Fatal and Panic.
	b.WriteString(strings.ToUpper(entry.Level.String()[:1]))
	b.WriteString(entry.Time.Format("0102 15:04:05.000000"))
	b.WriteString(" ")
	for i := 0; i < (7 - len(pidString)); i++ {
		b.WriteString(" ")
	}
	b.WriteString(pidString)
	b.WriteString(" ")
	b.WriteString(fileName)
	b.WriteString(":")
	fmt.Fprint(b, entry.Caller.Line)
	b.WriteString("] ")
	if entry.Message != "" {
		b.WriteString(entry.Message)
	}
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&klogFormatter{})
}
