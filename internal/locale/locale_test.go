package locale

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFallbackWhenKeyMissing(t *testing.T) {
	SetLanguage(LangEnglish)
	assert.Equal(t, "totally.unknown.key", T("totally.unknown.key"))
}

func TestLanguageSwitchAndFallback(t *testing.T) {
	SetLanguage(LangGerman)
	assert.Equal(t, "Installieren", T("action.install"))
	assert.Equal(t, "Pr!nt", T("app.name"))

	SetLanguage(LangGerman)
	assert.Equal(t, "Ich verstehe, schreibe auf %s", T("action.confirm"))  // present in de
	assert.Equal(t, "error.confirm_required", T("error.confirm_required")) // absent everywhere -> key returned

	SetLanguage(Language("xx"))
	assert.Equal(t, "Install", T("action.install"))
}

func TestFormatArgs(t *testing.T) {
	SetLanguage(LangEnglish)
	got := T("warn.destructive", "/dev/sda")
	assert.Equal(t, "This will erase all data on /dev/sda.", got)
}

func TestAvailableIncludesFallback(t *testing.T) {
	langs := Available()
	assert.Contains(t, langs, LangEnglish)
	assert.GreaterOrEqual(t, len(langs), 5)
}
