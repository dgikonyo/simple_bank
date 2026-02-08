package util

import "github.com/spf13/viper"

/**
 * Config - stores all configuration of the application.
 * @DBSource: The connection string for the database.
 * @ServerAddress: The host and port the server will run on.
 *
 * Description: Values are read by viper from a config file
 * or environment variables.
 */
type Config struct {
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

/**
 * LoadConfig - reads configuration from file or environment variables.
 * @path: The directory path containing the config file.
 *
 * Return: (config, err) The populated config struct and any error encountered.
 */
func LoadConfig(path string) (config Config, err error) {
	/* Tell viper where to look for the config file */
	viper.AddConfigPath(path)
	/* Look for a file named app (app.env) */
	viper.SetConfigName("app")
	/* Set the format of the config file */
	viper.SetConfigType("env")

	/* Allow environment variables to override file values */
	viper.AutomaticEnv()

	/* Start reading the config file */
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	/* Unmarshal the read values into the config struct */
	err = viper.Unmarshal(&config)
	return
}