package factories

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

func NewViperCfg(cfgPath string, prefix string) (cfg *viper.Viper, err error) {
	cfg = viper.New()
	cfg.SetEnvPrefix(prefix)
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	cfg.SetConfigFile(cfgPath)

	if err = cfg.ReadInConfig(); err != nil {
		return
	}

	// override settings if *.override.yml exists
	overrideCfgPath := strings.TrimSuffix(cfgPath, path.Ext(cfgPath)) + ".override" + path.Ext(cfgPath)
	if _, fsErr := os.Stat(overrideCfgPath); fsErr == nil {
		cfg.SetConfigFile(overrideCfgPath)
		err = cfg.MergeInConfig()
	}

	// workaround because viper does not treat env vars the same as other config
	for _, key := range cfg.AllKeys() {
		val := cfg.Get(key)
		cfg.Set(key, val)
	}

	return
}
