package config

import "flag"

var (
	flagServerAddr           NetAddress
	flagDSN                  string
	flagSecretKey            string
	flagAccrualSystemAddress string
)

func parseFlags() {
	flagServerAddr = NetAddress{Host: "localhost", Port: 8080}
	flag.Var(&flagServerAddr, "a", "address and port server")
	flag.StringVar(&flagDSN, "d", "", "database uri (ex. ")
	flag.StringVar(&flagSecretKey, "s", "DEBUG", "secret key for deploy")
	flag.StringVar(&flagAccrualSystemAddress, "r", "", "accrual system address")

	flag.Parse()
}
