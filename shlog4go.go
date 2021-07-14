package shlog4go

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type SHLogger struct {
	filename   string
	mutex      sync.Mutex
	prefix     string
	timeformat string
	out        *os.File
	categories map[string]int
	levelmap   map[string]int
	deflevel   int
}

func Open(filename string) (log *SHLogger, err error) {
	out, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	log = new(SHLogger)
	log.filename = filename
	log.out = out
	log.categories = make(map[string]int)
	log.SetLevelMap(map[string]int{
		"OFF":   1,
		"FATAL": 2,
		"ERROR": 3,
		"WARN":  4,
		"INFO":  5,
		"DEBUG": 6,
		"ALL":   7,
	})
	log.SetDefaultLevel("WARN")

	return log, nil
}

func (log *SHLogger) Close() {
	log.out.Close()
}

func (log *SHLogger) Reopen() error {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	log.Close()
	out, err := os.OpenFile(log.filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	log.out = out
	return nil
}

func (log *SHLogger) SetPrefix(prefix string) {
	log.prefix = prefix
}

func (log *SHLogger) SetTimeFormat(timeformat string) {
	log.timeformat = timeformat
}

func (log *SHLogger) SetLevelMap(lm map[string]int) {
	log.levelmap = lm
}

func (log *SHLogger) SetDefaultLevel(level string) {
	log.deflevel = log.levelmap[level]
}

func (log *SHLogger) SetCategory(category string, level string) {
	log.categories[category] = log.levelmap[level]
}

func getShortFileName(filename string) string {
	return filename[strings.LastIndex(filename, "/")+1:]
}

func getTimeString(timeformat string) string {
	now := time.Now()
	if timeformat == "" {
		timeformat = time.RFC3339
	}
	return now.Format(timeformat)
}

func (log *SHLogger) formatHeader(category string, level string, buf *[]byte) {
	pc, file, line, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	end := len(log.prefix)
	for i := 0; i < end; {
		lasti := i
		for i < end && log.prefix[i] != '%' {
			i++
		}
		if i > lasti {
			*buf = append(*buf, log.prefix[lasti:i]...)
		}
		i++
		if i >= end {
			break
		}
		switch log.prefix[i] {
		case 'p':
			*buf = append(*buf, fmt.Sprintf("%d", pc)...)
		case 'F':
			*buf = append(*buf, file...)
		case 'f':
			*buf = append(*buf, getShortFileName(file)...)
		case 'l':
			*buf = append(*buf, fmt.Sprintf("%d", line)...)
		case 'm':
			*buf = append(*buf, f.Name()...)
		case 't':
			*buf = append(*buf, getTimeString(log.timeformat)...)
		case 'c':
			*buf = append(*buf, category...)
		case 'L':
			*buf = append(*buf, level...)
		case '%':
			*buf = append(*buf, "%"...)
		default:
		}
		i++
	}
}

func (log *SHLogger) checkPrintable(category string, level string) bool {
	lv := log.levelmap[level]
	cl, ok := log.categories[category]
	if ok {
		return lv <= cl
	} else {
		return lv <= log.deflevel
	}
}

func (log *SHLogger) Printf(category string, level string, format string, a ...interface{}) (n int, err error) {
	if !log.checkPrintable(category, level) {
		return
	}
	var buf []byte
	log.formatHeader(category, level, &buf)
	buf = append(buf, fmt.Sprintf(format, a...)...)
	log.mutex.Lock()
	defer log.mutex.Unlock()
	return log.out.Write(buf)
}

func (log *SHLogger) Println(category string, level string, a ...interface{}) (n int, err error) {
	if !log.checkPrintable(category, level) {
		return
	}
	var buf []byte
	log.formatHeader(category, level, &buf)
	buf = append(buf, fmt.Sprintln(a...)...)
	log.mutex.Lock()
	defer log.mutex.Unlock()
	return log.out.Write(buf)
}

func (log *SHLogger) Sprintf(category string, level string, format string, a ...interface{}) string {
	if !log.checkPrintable(category, level) {
		return ""
	}
	var buf []byte
	log.formatHeader(category, level, &buf)
	buf = append(buf, fmt.Sprintf(format, a...)...)
	return string(buf)
}

func (log *SHLogger) Sprintln(category string, level string, a ...interface{}) string {
	if !log.checkPrintable(category, level) {
		return ""
	}
	var buf []byte
	log.formatHeader(category, level, &buf)
	buf = append(buf, fmt.Sprintln(a...)...)
	return string(buf)
}
