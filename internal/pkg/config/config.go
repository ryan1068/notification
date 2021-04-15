package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type application struct {
	Env  string
	Host string
	Port string
}

type mysqldb struct {
	IP       string
	Port     string
	Username string
	Password string
	Database string
}

type mongodb struct {
	IP       string
	Port     string
	Username string
	Password string
	Options  string
	Database string
}

type versions struct {
	Url string
}

type Config struct {
	Application application
	MysqlDb     mysqldb
	MongoDb     mongodb
	Versions    versions
}

func Load(file string) (*Config, error) {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Fatal error configs file: %s \n", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}
	return cfg, nil
}

func (db *mysqldb) Dsn() string {
	return db.Username + ":" + db.Password + "@tcp(" + db.IP + ":" + db.Port + ")/" + db.Database + "?parseTime=true"
}

func (db *mongodb) Dsn() string {
	dsn := "mongodb://"
	if len(db.Username) > 0 && len(db.Password) > 0 {
		dsn = dsn + db.Username + ":" + db.Password + "@"
	}

	dsn = dsn + db.IP + ":" + db.Port

	if len(db.Options) > 0 {
		dsn = dsn + "?" + db.Options
	}

	return dsn
}
