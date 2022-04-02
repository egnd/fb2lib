package factories

import (
	"strings"

	"github.com/spf13/viper"
)

func NewViperCfg(path string, prefix string) (cfg *viper.Viper, err error) {
	cfg = viper.New()
	cfg.SetEnvPrefix(prefix)
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	cfg.SetConfigFile(path)
	err = cfg.ReadInConfig()

	return
}
