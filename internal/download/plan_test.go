package download

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

type fakeRepository struct {
	languages []string
	exported  []File
	failAt    int
}

func (repo *fakeRepository) Languages(context.Context, Project) ([]string, error) {
	return repo.languages, nil
}
func (repo *fakeRepository) Export(_ context.Context, _ Project, file File) ([]byte, error) {
	repo.exported = append(repo.exported, file)
	if len(repo.exported) == repo.failAt {
		return nil, errors.New("download failed")
	}
	return []byte("content"), nil
}

type fakeStore struct{ committed []DownloadedFile }

func (*fakeStore) Validate([]File) error                     { return nil }
func (store *fakeStore) Commit(files []DownloadedFile) error { store.committed = files; return nil }

func config() Config {
	return Config{Directory: filepath.Join(string(filepath.Separator), "repo"), Projects: map[string]Project{"app": {ID: "p"}},
		Outputs: map[string]Output{"ios": {Project: "app", Format: "xcstrings", Layout: Combined, Languages: []string{"en", "cs"}, SourceLanguage: "en", Path: "Localizable.xcstrings"}}}
}

func TestCombinedPlanProducesOneFile(t *testing.T) {
	files, err := Plan(context.Background(), config(), nil, &fakeRepository{})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].SourceLanguage != "en" || !reflect.DeepEqual(files[0].Languages, []string{"cs", "en"}) {
		t.Fatalf("incorrect combined plan: %+v", files)
	}
}

func TestPerLanguagePathsUseExactAndMappedLanguages(t *testing.T) {
	config := config()
	config.Outputs["ios"] = Output{Project: "app", Format: "apple", Layout: PerLanguage, Languages: []string{"en_US", "cs_CZ"},
		Path: "{locale}/{language}.strings", LocaleMap: map[string]string{"en_US": "en", "cs_CZ": "cs"}}
	files, err := Plan(context.Background(), config, nil, &fakeRepository{})
	if err != nil {
		t.Fatal(err)
	}
	if files[0].Path != filepath.Join(config.Directory, "cs", "cs_CZ.strings") {
		t.Fatalf("wrong resolved path: %q", files[0].Path)
	}
	if files[0].Languages[0] != "cs_CZ" {
		t.Fatal("path alias changed API language")
	}
}

func TestAllLanguagesRequiresCompleteLocaleMap(t *testing.T) {
	config := config()
	config.Outputs["ios"] = Output{Project: "app", Format: "apple", Layout: PerLanguage, AllLanguages: true, Path: "{locale}.strings", LocaleMap: map[string]string{"en": "en"}}
	if _, err := Plan(context.Background(), config, nil, &fakeRepository{languages: []string{"en", "cs"}}); err == nil {
		t.Fatal("accepted missing locale mapping")
	}
}

func TestCombinedSourceMustBeSelected(t *testing.T) {
	config := config()
	output := config.Outputs["ios"]
	output.SourceLanguage = "de"
	config.Outputs["ios"] = output
	if _, err := Plan(context.Background(), config, nil, &fakeRepository{}); err == nil {
		t.Fatal("accepted unselected source")
	}
}

func TestDestinationCollisionsAreRejected(t *testing.T) {
	config := config()
	config.Outputs["other"] = config.Outputs["ios"]
	if _, err := Plan(context.Background(), config, nil, &fakeRepository{}); err == nil {
		t.Fatal("accepted conflicting outputs")
	}
}

func TestCaseInsensitiveDestinationCollisionsAreRejected(t *testing.T) {
	config := config()
	other := config.Outputs["ios"]
	other.Path = "LOCALIZABLE.XCSTRINGS"
	config.Outputs["other"] = other
	if _, err := Plan(context.Background(), config, nil, &fakeRepository{}); err == nil {
		t.Fatal("accepted case-insensitive collision")
	}
}

func TestDestinationCannotBeAnotherOutputsDirectory(t *testing.T) {
	config := config()
	other := config.Outputs["ios"]
	other.Path = "Localizable.xcstrings/child"
	config.Outputs["other"] = other
	if _, err := Plan(context.Background(), config, nil, &fakeRepository{}); err == nil {
		t.Fatal("accepted file/directory collision")
	}
}

func TestDownloadFailureDoesNotCommitAnyFiles(t *testing.T) {
	store := &fakeStore{}
	repo := &fakeRepository{failAt: 2}
	files := []File{{Output: "first", Project: "app"}, {Output: "second", Project: "app"}}
	if err := Pull(context.Background(), config(), files, repo, store); err == nil {
		t.Fatal("download error was lost")
	}
	if store.committed != nil {
		t.Fatal("partial download was committed")
	}
}

func TestLanguageAliases(t *testing.T) {
	cases := []struct {
		name      string
		languages []string
		all       bool
		aliases   map[string]string
		want      map[string]string
		invalid   bool
	}{
		{name: "unselected source", languages: []string{"sk-SK"}, aliases: map[string]string{"sk-SK": "cs-CZ"}, want: map[string]string{"sk-SK": "cs-CZ"}},
		{name: "chain", languages: []string{"sk-SK"}, aliases: map[string]string{"sk-SK": "alias", "alias": "cs-CZ"}, want: map[string]string{"sk-SK": "cs-CZ"}},
		{name: "override actual", languages: []string{"en"}, aliases: map[string]string{"en": "cs-CZ"}, want: map[string]string{"en": "cs-CZ"}},
		{name: "wildcard", all: true, aliases: map[string]string{"sk-SK": "cs-CZ"}, want: map[string]string{"en": "en", "cs-CZ": "cs-CZ", "sk-SK": "cs-CZ"}},
		{name: "cycle", languages: []string{"sk-SK"}, aliases: map[string]string{"sk-SK": "en", "en": "sk-SK"}, invalid: true},
		{name: "missing source", languages: []string{"sk-SK"}, aliases: map[string]string{"sk-SK": "missing"}, invalid: true},
		{name: "unselected invalid alias", languages: []string{"en"}, aliases: map[string]string{"sk-SK": "missing"}, invalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config()
			cfg.Outputs = map[string]Output{"mobile": {Project: "app", Format: "mo", Layout: PerLanguage, Languages: tc.languages, AllLanguages: tc.all, LanguageAliases: tc.aliases, Path: "{language}.mo"}}
			files, err := Plan(context.Background(), cfg, nil, &fakeRepository{languages: []string{"en", "cs-CZ"}})
			if tc.invalid {
				if err == nil {
					t.Fatal("accepted invalid aliases")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]string{}
			for _, file := range files {
				got[filepath.Base(file.Path)[:len(filepath.Base(file.Path))-3]] = file.Languages[0]
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
