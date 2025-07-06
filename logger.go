package main

import (
    "fmt"
    "regexp"

	waLog "go.mau.fi/whatsmeow/util/log"
)

type RIVAClientLog struct {
    logger   waLog.Logger
    level    string
}

func NewRIVAClientLog(name string, logLevel string) *RIVAClientLog {
    return &RIVAClientLog{
        logger: waLog.Stdout(name, logLevel, true),
        level:  logLevel,
    }
}

// No redaction for debug
func (l *RIVAClientLog) Debugf(format string, v ...interface{}) {
    l.logger.Debugf(format, v...)
}

func (l *RIVAClientLog) Infof(format string, v ...interface{}) {
    l.logger.Infof(l.redact(format, v...))
}

func (l *RIVAClientLog) Warnf(format string, v ...interface{}) {
    l.logger.Warnf(l.redact(format, v...))
}

func (l *RIVAClientLog) Errorf(format string, v ...interface{}) {
    l.logger.Errorf(l.redact(format, v...))
}

func (l *RIVAClientLog) redact(format string, v ...interface{}) string {
    raw := fmt.Sprintf(format, v...)

    /*
     * Usually the phone numbers logged for private messages would be in this
     * format:
     * 
     *   From: 6581234567@s.whatsapp.net
     *
     * We would need to redact the last 4-digits of the phone number and thus,
     * we create a capture group for the first 6-digits and ignore the rest:
     *
     *   Capture Group: 658123
     *   Ignored Numbers: 4567
     */
    regex := regexp.MustCompile(`(\d{6,})\d{4}`)
    redacted := regex.ReplaceAllString(raw, "${1}XXXX")

    return redacted
}

