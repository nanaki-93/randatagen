package config

import (
	"fmt"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/spf13/viper"
)

var DynamicQueries model.DynamicQueries

func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name ".randatagen" (without extension).
		viper.AddConfigPath("./")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".randatagen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = viper.Unmarshal(&DynamicQueries)

	if err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}
	return nil
}
