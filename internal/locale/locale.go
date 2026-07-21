package locale

import (
	"strings"
	"sync"
)

type Language string

const (
	LangEnglish  Language = "en"
	LangGerman   Language = "de"
	LangSpanish  Language = "es"
	LangJapanese Language = "ja"
	LangFrench   Language = "fr"
)

const fallback = LangEnglish

var translations = map[Language]map[string]string{
	LangEnglish:  en,
	LangGerman:   de,
	LangSpanish:  es,
	LangJapanese: ja,
	LangFrench:   fr,
}

type translator struct {
	lang Language
	mu   sync.RWMutex
}

var (
	current *translator
	once    sync.Once
)

func init() {
	current = &translator{lang: LangEnglish}
}

func SetLanguage(l Language) {
	once.Do(func() { current = &translator{lang: LangEnglish} })
	current.mu.Lock()
	defer current.mu.Unlock()
	if _, ok := translations[normalize(l)]; !ok {
		current.lang = fallback
		return
	}
	current.lang = normalize(l)
}

func T(key string, args ...any) string {
	t := current
	t.mu.RLock()
	lang := t.lang
	t.mu.RUnlock()

	if m, ok := translations[lang]; ok {
		if s, ok := m[key]; ok {
			return format(s, args...)
		}
	}
	if m, ok := translations[fallback]; ok {
		if s, ok := m[key]; ok {
			return format(s, args...)
		}
	}
	return key
}

func Available() []Language {
	out := make([]Language, 0, len(translations))
	for l := range translations {
		out = append(out, l)
	}
	return out
}

func normalize(l Language) Language {
	return Language(strings.ToLower(strings.TrimSpace(string(l))))
}

func format(s string, args ...any) string {
	if len(args) == 0 {
		return s
	}
	return sprintf(s, args...)
}
