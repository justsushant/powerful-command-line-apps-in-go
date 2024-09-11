package notify

import (
	"runtime"
	"strings"
)

const (
	SeverityLow = iota
	SeverityNormal
	SeverityUrgent
)

// a custom type Severity with an underlying type int to represent the severity
// by doing this, methods can be attached to this type
// like, the method String that returns the severity’s string representation 
type Severity int

type Notify struct {
	title string
	message string
	severity Severity
}

func New(title, message string, severity Severity) *Notify {
	return &Notify{
		title: title,
		message: message,
		severity: severity,
	}
}

func (s Severity) String() string {
	var sev string

	switch s {
	case SeverityLow:
		sev = "low"
	case SeverityNormal:
		sev = "normal"
	case SeverityUrgent:
		sev = "critical"
	}

	if runtime.GOOS == "darwin" {
		sev = strings.Title(sev)
	}

	if runtime.GOOS == "windows" {
		switch s {
		case SeverityLow:
			sev = "Info"
		case SeverityNormal:
			sev = "Warning"
		case SeverityUrgent:
			sev = "Error"
		}
	}

	return sev
}

