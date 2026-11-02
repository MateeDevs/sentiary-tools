package download

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestLegacyRequestIdentityRemainsByteCompatible(t *testing.T) {
	file := File{Project: "app", Format: "json", Layout: PerLanguage, Languages: []string{"en"}}
	request := RequestFor(config(), file)
	legacy := []byte(`{"ProjectID":"p","Format":"json","Layout":"per_language","Languages":["en"],"SourceLanguage":"","IncludeEmpty":false}`)
	if request.ID() != Digest(legacy) {
		data, _ := json.Marshal(request)
		t.Fatalf("legacy identity changed: %s", data)
	}
}

func TestExportIdentityUsesCanonicalContent(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{"b", "a"}
	file := File{Project: "app", Format: "android", Layout: PerLanguage, Languages: []string{"en"}, Options: ExportOptions{Keys: &a, RequiredKeys: []string{"a", "b"}}}
	first := RequestFor(config(), file).ID()
	file.Options = ExportOptions{Keys: &b, RequiredKeys: []string{"b", "a"}}
	if first != RequestFor(config(), file).ID() {
		t.Fatal("order changed identity")
	}
	if !reflect.DeepEqual(b, []string{"b", "a"}) {
		t.Fatal("canonicalization mutated input")
	}
	file.Options.RequiredKeys = nil
	if first == RequestFor(config(), file).ID() {
		t.Fatal("required policy not in identity")
	}
	file.Options.PluralBases = []string{"count"}
	plural := RequestFor(config(), file).ID()
	file.Options.PluralBases = nil
	if plural == RequestFor(config(), file).ID() {
		t.Fatal("plurals not in identity")
	}
	file.Options.Contexts = []string{"screen"}
	contextID := RequestFor(config(), file).ID()
	file.Options.Contexts = nil
	if contextID == RequestFor(config(), file).ID() {
		t.Fatal("contexts not in identity")
	}
	empty := []string{}
	file.Options = ExportOptions{Keys: &empty}
	explicit := RequestFor(config(), file).ID()
	file.Options = ExportOptions{}
	if explicit == RequestFor(config(), file).ID() {
		t.Fatal("explicit empty selection aliases full catalog")
	}
}

func TestResourceSnapshotRestoresExactBytesAndRejectsDifferentPolicy(t *testing.T) {
	cache := &memoryCache{}
	repo := newSnapshotRepo()
	repo.state.Capabilities = []string{ResourceExportCapability}
	cfg := config()
	out := cfg.Outputs["ios"]
	keys := []string{"a", "b"}
	out.Options = ExportOptions{Keys: &keys, RequiredKeys: []string{"a"}}
	cfg.Outputs["ios"] = out
	files, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache)
	if err != nil {
		t.Fatal(err)
	}
	keys = []string{"b", "a"}
	out.Options.Keys = &keys
	cfg.Outputs["ios"] = out
	restored, _, err := RestoreSnapshot(context.Background(), cfg, nil, cache)
	if err != nil || !reflect.DeepEqual(files[0].Content, restored[0].Content) {
		t.Fatalf("bytes changed: %v", err)
	}
	out.Options.RequiredKeys = nil
	cfg.Outputs["ios"] = out
	if _, _, err := RestoreSnapshot(context.Background(), cfg, nil, cache); err == nil {
		t.Fatal("restored mismatched policy")
	}
}

func TestOldServerCannotServeResourceOptions(t *testing.T) {
	cfg := config()
	out := cfg.Outputs["ios"]
	out.Options.Contexts = []string{"screen"}
	cfg.Outputs["ios"] = out
	repo := newSnapshotRepo()
	cache := &memoryCache{}
	if _, err := FetchSnapshot(context.Background(), cfg, nil, repo, cache); err == nil {
		t.Fatal("accepted old server")
	}
	if repo.calls != 0 || len(cache.sets) != 0 {
		t.Fatal("exported or published unsupported request")
	}
}
