package config

import "testing"

func TestDigestLangPrefersDigestLang(t *testing.T) {
	t.Setenv("DIGEST_LANG", "en")
	t.Setenv("LANG", "C.UTF-8")

	got, err := digestLang()
	if err != nil {
		t.Fatalf("digestLang() error = %v", err)
	}
	if got != "en" {
		t.Fatalf("digestLang() = %q, want %q", got, "en")
	}
}

func TestDigestLangIgnoresLocaleLang(t *testing.T) {
	t.Setenv("LANG", "C.UTF-8")

	got, err := digestLang()
	if err != nil {
		t.Fatalf("digestLang() error = %v", err)
	}
	if got != "bilingual" {
		t.Fatalf("digestLang() = %q, want %q", got, "bilingual")
	}
}

func TestDigestLangAllowsLegacyLang(t *testing.T) {
	t.Setenv("LANG", "pt")

	got, err := digestLang()
	if err != nil {
		t.Fatalf("digestLang() error = %v", err)
	}
	if got != "pt" {
		t.Fatalf("digestLang() = %q, want %q", got, "pt")
	}
}

func TestDigestLangRejectsInvalidDigestLang(t *testing.T) {
	t.Setenv("DIGEST_LANG", "fr")

	if _, err := digestLang(); err == nil {
		t.Fatal("digestLang() error = nil, want error")
	}
}
