package environment

func PostgresConnectionString(envType string) (string) {
	return "user=postgres dbname=88mph_development sslmode=disable"
}