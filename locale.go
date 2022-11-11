package locale

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// Singleton instance of i18n.Bundle
var bundle *i18n.Bundle

// A list of localizers for different languages
var localizers = make(map[language.Tag]*i18n.Localizer, 0)

// Fallback language to be used when no suitable
// localization string is not found
var fallbackLanguage language.Tag

var mocking = false

func makeBundle(defaultLanguage language.Tag) *i18n.Bundle {
	bundle = i18n.NewBundle(defaultLanguage)

	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	return bundle
}

// Load localization files from directory specified by the <directoryPath>
// and prepare them for global usage.
//
//  //go:embed locale/*.yml
//  var localeFS embed.FS
//  err := locale.Initialize(languages.English, localeFS, "locale")
func Initialize(
	defaultLanguage language.Tag,
	filesystem fs.ReadDirFS,
	directoryPath string,
) error {
	fallbackLanguage = defaultLanguage
	bundle = makeBundle(defaultLanguage)

	if directoryPath == "" {
		directoryPath = "."
	}

	// Read all the files from the directory
	entries, err := filesystem.ReadDir(directoryPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Load each file into the bundle
		_, err = bundle.LoadMessageFileFS(
			filesystem,
			directoryPath+"/"+entry.Name(),
		)
		if err != nil {
			return err
		}

		// Deduce language from the file name (without extension)
		tag := language.Make(
			strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())),
		)

		localizers[tag] = i18n.NewLocalizer(bundle, tag.String())
	}

	return nil
}

func InitializeMock() error {
	defaultLanguage := language.English

	fallbackLanguage = defaultLanguage
	bundle = makeBundle(defaultLanguage)

	localizers[defaultLanguage] = i18n.NewLocalizer(
		bundle,
		defaultLanguage.String(),
	)

	mocking = true

	return nil
}

func GetLanguages() []language.Tag {
	checkIfInitializaed()

	keys := []language.Tag{}

	for k := range localizers {
		keys = append(keys, k)
	}

	return keys
}

func GetDefaultLanguage() language.Tag {
	checkIfInitializaed()

	return fallbackLanguage
}

func GetMessage(id string, tag language.Tag, args []any, count interface{}, fallback string) string {
	data := parseArgs(args)
	if data["Count"] == nil {
		data["Count"] = count
	}

	if fallback == "" {
		fallback = id
	}

	message, _ := getLocalizer(tag).Localize(&i18n.LocalizeConfig{
		TemplateData: data,
		PluralCount:  count,
		DefaultMessage: &i18n.Message{
			ID:    id,
			Other: fallback,
		},
	})

	return message
}
