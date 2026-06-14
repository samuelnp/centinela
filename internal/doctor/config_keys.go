package doctor

import (
	"os"
	"sort"

	"github.com/BurntSushi/toml"

	"github.com/samuelnp/centinela/internal/config"
)

// unknownConfigKeys re-decodes centinela.toml against the config schema and
// reports any keys present in the file but not recognized by the schema. It
// uses BurntSushi's MetaData.Undecoded(), which lists exactly the keys the
// decoder did not map onto a struct field. A missing or unparseable file yields
// no findings (the parse-error path is handled by the config check itself).
func unknownConfigKeys() []string {
	data, err := os.ReadFile(config.Filename)
	if err != nil {
		return nil
	}
	var cfg config.Config
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return nil
	}
	var out []string
	for _, key := range md.Undecoded() {
		out = append(out, "unknown config key: "+key.String())
	}
	sort.Strings(out)
	return out
}
