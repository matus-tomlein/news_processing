package environment

func PostgresConnectionString(envType string) (string) {
	if envType == "production" {
		return "user=postgres dbname=88mph_production password=noroads sslmode=disable"
	} else if envType == "calculon-test" {
		return "user=postgres dbname=88mph_test password=noroads sslmode=disable"
	}
	return "user=postgres dbname=88mph_development sslmode=disable"
}
