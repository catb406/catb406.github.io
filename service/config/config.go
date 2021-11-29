package config

import "time"

var RefreshTokenExpiration = 30 * time.Hour * 24
var AccessTokenExpiration = 60 * time.Minute

var DbConnectTimeout = 5 * time.Second

const (
	PostgresUser     = "postgres"
	PostgresPassword = "1"
	PostgresPort     = "5432"
	PostgresDbName   = "sport_buddy_v0"

	Secret = "s0Wo!GLNLkjwVG4G:Jf18/KvAM"
)
