package opentracer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

type Logger interface {
	LogFields(fields ...log.Field)
	LogFieldsTime(t time.Time, fields ...log.Field)
}

var (
	Stdout Logger = writeLogger{w: os.Stdout}
)

type writeLogger struct {
	w io.Writer
}

func (s writeLogger) LogFields(fields ...log.Field) {
	s.LogFieldsTime(time.Now(), fields...)
}

func (s writeLogger) LogFieldsTime(t time.Time, fields ...log.Field) {
	var buffer = bytes.NewBuffer(nil)

	buffer.WriteString(t.Format(time.RFC822Z))
	buffer.WriteString(" ")

	for index, field := range fields {
		if index > 0 {
			buffer.WriteString(" ")
		}
		fmt.Fprintf(buffer, "%v=%v", field.Key(), field.Value())
	}
	fmt.Fprintln(s.w, buffer.String())
}
