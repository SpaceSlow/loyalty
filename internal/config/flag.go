package config

import "flag"

var (
	flagDSN       string
	flagSecretKey string
)

func parseFlags() {
	flag.StringVar(&flagDSN, "d", "", "database uri (ex. ")
	flag.StringVar(&flagSecretKey, "s", "DEBUG", "secret key for deploy")

	flag.Parse()
}
