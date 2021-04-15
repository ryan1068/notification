package db

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"notification/internal/pkg/config"
)

type mongodb struct {
	cfg    *config.Config
	logger *logrus.Logger
}

func NewMongoDB(cfg *config.Config, logger *logrus.Logger) *mongodb {
	return &mongodb{
		logger: logger,
		cfg:    cfg,
	}
}

// Opening a database and save the reference to `Database` struct.
func (m *mongodb) Open() (*mgo.Database, error) {
	session, err := mgo.Dial(m.cfg.MongoDb.Dsn())
	if err != nil {
		m.logger.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Error(fmt.Sprintf("mongodb open error, err:%s", err.Error()))
		return nil, err
	}

	return session.DB(m.cfg.MongoDb.Database), nil
}
