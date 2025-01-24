package main

import (
	"regexp"
	"strings"
	"testing"
)

func BenchmarkStringsSplit(b *testing.B) {
	text := "bla.bla.GET"
	for i := 0; i < b.N; i++ {
		parts := strings.Split(text, ".")
		res := parts[len(parts)-1]
		if res != "GET" {
			b.Fail()
		}
	}
}

func BenchmarkRegexpSplit(b *testing.B) {
	text := "bla.bla.GET"
	re := regexp.MustCompile(`\.(\w+)$`)
	for i := 0; i < b.N; i++ {
		res := re.FindStringSubmatch(text)
		if res[1] != "GET" {
			b.Fail()
		}
	}
}

func ByteSplit(s string) string {
	var lastDot int
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			lastDot = i
		}
	}
	return s[lastDot+1:]
}

func BenchmarkByteSplit(b *testing.B) {
	text := "bla.bla.GET"
	for i := 0; i < b.N; i++ {
		res := ByteSplit(text)
		if res != "GET" {
			b.Fail()
		}

	}
}
