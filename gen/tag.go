package main

import (
	"log"
	"reflect"
	"strings"
)

type tag struct {
	Key      string
	Mutators []mutator
}

func (t tag) HasDive() bool {
	log.Print(t.Mutators)
	for _, mu := range t.Mutators {
		if mu.Name == DiveDirective {
			return true
		}
	}

	return false
}

type mutator struct {
	Name string
	Args map[string]string
}

type tagParser struct {
	tag       tag
	mu        mutator
	key       strings.Builder
	value     strings.Builder
	inQuote   bool
	escaped   bool
	hasEquals bool
}

func newTagParser() *tagParser {
	return &tagParser{
		mu: mutator{
			Args: make(map[string]string),
		},
	}
}

func (p *tagParser) flushArg() {
	if k := strings.TrimSpace(p.key.String()); k != "" {
		p.mu.Args[k] = p.value.String()
	}

	p.key.Reset()
	p.value.Reset()
	p.hasEquals = false
}

func (p *tagParser) flushMutator() {
	if !p.hasEquals && p.mu.Name == "" {
		p.mu.Name = p.key.String()
		p.key.Reset()
	} else {
		p.flushArg()
	}

	if p.mu.Name != "" {
		p.tag.Mutators = append(p.tag.Mutators, p.mu)
	}

	p.mu = mutator{Args: make(map[string]string)}
}

func (p *tagParser) consume(r rune) {
	if p.hasEquals {
		p.value.WriteRune(r)
	} else {
		p.key.WriteRune(r)
	}
}

func (p *tagParser) flushGroup() {
	if p.tag.Key != "" {
		p.flushMutator()
		return
	}

	p.tag.Key = strings.TrimSpace(p.key.String())
	p.key.Reset()
}

func (p *tagParser) Parse(in string) tag {
	st := reflect.StructTag(strings.Trim(in, "`"))

	for _, r := range st {
		if p.escaped {
			p.consume(r)
			p.escaped = false
			continue
		} else if r == '\\' && p.inQuote {
			p.escaped = true
			continue
		} else if r == '\'' {
			p.inQuote = !p.inQuote
			continue
		}

		if !p.inQuote {
			switch r {
			case ';':
				p.flushGroup()
				continue

			case ',':
				if p.hasEquals {
					p.flushArg()
				} else if p.mu.Name == "" {
					p.mu.Name = strings.TrimSpace(p.key.String())
					p.key.Reset()
				}
				continue

			case '=':
				if !p.hasEquals {
					p.hasEquals = true
					continue
				}
			}
		}

		p.consume(r)
	}

	p.flushGroup()

	return p.tag
}
