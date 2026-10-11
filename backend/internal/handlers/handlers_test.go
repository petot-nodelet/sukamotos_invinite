package handlers

import (
	"strings"
	"testing"
)

func TestReadTemplateSourceLoadsAllRegisteredOriginalTemplates(t *testing.T) {
	sources := []string{
		"templates/SILVER/template1_classic_matcha.html",
		"templates/SILVER/template4-burgundy-minimalist.html",
		"templates/SILVER/template8_islamic_traditional.html",
		"templates/GOLD/dusty_rose_wedding.html",
		"templates/GOLD/black-champagne-gold-ivory.html",
		"templates/PLATINUM/3_Imperial-Modern.html",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			page, err := readTemplateSource(source)
			if err != nil {
				t.Fatalf("readTemplateSource(%q): %v", source, err)
			}
			if !strings.Contains(page, "<html") || !strings.Contains(page, "/templates/personalize.js") {
				t.Fatalf("template %q is missing HTML or its personalization runtime", source)
			}
		})
	}
}

func TestReadTemplateSourceRejectsPathTraversal(t *testing.T) {
	if _, err := readTemplateSource("templates/../../etc/passwd"); err == nil {
		t.Fatal("readTemplateSource accepted a path outside the template directory")
	}
}
