package gosnmplib

import (
	"regexp"

	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Replacer structure to store regex matching logs parts to replace
type Replacer struct {
	Regex *regexp.Regexp
	Repl  []byte
}

// TODO: Test TraceLevelLogWriter replacements against real GoSNMP library output
//       (need more complex setup e.g. simulate gosnmp request/response)

var replacers = []Replacer{
	{
		Regex: regexp.MustCompile(`(\s*SECURITY PARAMETERS\s*:).+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*Parsed (privacyParameters|contextEngineID))\s*.+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(AuthenticationPassphrase|PrivacyPassphrase|SecretKey|PrivacyKey|authenticationParameters)\s*:).+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(authenticationParameters))\s*.+`),
		Repl:  []byte(`$1 ********`),
	},
	{
		Regex: regexp.MustCompile(`(\s*(?:Community|ContextEngineID):).+?(\s[\w]+:)`),
		Repl:  []byte(`${1}********${2}`),
	},
}

// TraceLevelLogWriter is a log writer for gosnmp logs, it removes sensitive info
type TraceLevelLogWriter struct{}

func (sw *TraceLevelLogWriter) Write(logInput []byte) (n int, err error) {
	for _, replacer := range replacers {
		if replacer.Regex.Match(logInput) {
			logInput = replacer.Regex.ReplaceAll(logInput, replacer.Repl)
		}
	}
	log.Tracef(string(logInput))
	return len(logInput), nil
}
