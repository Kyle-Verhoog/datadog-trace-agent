// Package quantizer implements quantizing and obfuscating of tags and resources for
// a set of spans matching a certain criteria.
package quantizer

import (
	"github.com/DataDog/datadog-trace-agent/config"
	"github.com/DataDog/datadog-trace-agent/model"
)

// Obfuscator quantizes and obfuscates spans.
type Obfuscator struct {
	opts  *config.ObfuscationConfig
	es    *jsonObfuscator // nil if disabled
	mongo *jsonObfuscator // nil if disabled
}

// NewObfuscator creates a new Obfuscator.
func NewObfuscator(cfg *config.ObfuscationConfig) *Obfuscator {
	if cfg == nil {
		cfg = new(config.ObfuscationConfig)
	}
	o := Obfuscator{opts: cfg}
	if cfg.ES.Enabled {
		o.es = newJSONObfuscator(&cfg.ES)
	}
	if cfg.Mongo.Enabled {
		o.mongo = newJSONObfuscator(&cfg.Mongo)
	}
	return &o
}

// Obfuscate may obfuscate span's properties based on its type and on the Obfuscator's
// configuration.
func (o *Obfuscator) Obfuscate(span *model.Span) {
	switch span.Type {
	case "sql", "cassandra":
		o.quantizeSQL(span)
	case "redis":
		o.quantizeRedis(span)
	case "mongodb":
		o.obfuscateJSON(span, "mongodb.query", o.mongo)
	case "elasticsearch":
		o.obfuscateJSON(span, "elasticsearch.body", o.es)
	}
}

// compactWhitespaces compacts all whitespaces in t.
func compactWhitespaces(t string) string {
	n := len(t)
	r := make([]byte, n)
	spaceCode := uint8(32)
	isWhitespace := func(char uint8) bool { return char == spaceCode }
	nr := 0
	offset := 0
	for i := 0; i < n; i++ {
		if isWhitespace(t[i]) {
			copy(r[nr:], t[nr+offset:i])
			r[i-offset] = spaceCode
			nr = i + 1 - offset
			for j := i + 1; j < n; j++ {
				if !isWhitespace(t[j]) {
					offset += j - i - 1
					i = j
					break
				} else if j == n-1 {
					offset += j - i
					i = j
					break
				}
			}
		}
	}
	copy(r[nr:], t[nr+offset:n])

	// Trim
	rStart := 0
	rEnd := n - offset
	if isWhitespace(r[rEnd-1]) {
		rEnd--
	}
	if isWhitespace(r[rStart]) {
		rStart++
	}

	return string(r[rStart:rEnd])
}
