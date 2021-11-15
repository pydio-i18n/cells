package cmd

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// bindViperFlags visits all flags in FlagSet and bind their key to the corresponding viper variable
func bindViperFlags(flags *pflag.FlagSet, replaceKeys map[string]string) {
	flags.VisitAll(func(flag *pflag.Flag) {
		key := flag.Name
		if replace, ok := replaceKeys[flag.Name]; ok {
			key = replace
		}
		viper.BindPFlag(key, flag)
	})
}
