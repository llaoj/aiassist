package i18n

import "fmt"

// Messages maps language codes to message dictionaries
var Messages = map[string]map[string]string{
	"en": EnglishMessages,
	"zh": ChineseMessages,
}

// I18n manages internationalization
type I18n struct {
	language string
}

// New creates a new I18n instance
func New(language string) *I18n {
	return &I18n{language: language}
}

// T translates a message key with optional arguments
func (i *I18n) T(key string, args ...interface{}) string {
	msgs, ok := Messages[i.language]
	if !ok {
		msgs = Messages["en"]
	}

	msg, ok := msgs[key]
	if !ok {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func (i *I18n) SetLanguage(language string) {
	i.language = language
}

func (i *I18n) GetLanguage() string {
	return i.language
}
