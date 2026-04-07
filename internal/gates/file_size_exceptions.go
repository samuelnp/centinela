package gates

import "github.com/samuelnp/centinela/internal/config"

func fileSizeExceptionMap(cfg *config.Config) map[string]config.FileSizeException {
	out := map[string]config.FileSizeException{}
	if cfg == nil {
		return out
	}
	for _, ex := range cfg.Gates.FileSizeExceptions {
		out[ex.Path] = ex
	}
	return out
}
