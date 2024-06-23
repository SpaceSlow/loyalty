package config

import "flag"

var (
	flagDSN                  string
	flagSecretKey            string
	flagAccrualSystemAddress string
)

func parseFlags() {
	flag.StringVar(&flagDSN, "d", "", "database uri (ex. ")
	flag.StringVar(&flagSecretKey, "s", "DEBUG", "secret key for deploy")
	flag.StringVar(&flagAccrualSystemAddress, "r", "", "accrual system address")

	flag.Parse()
}
