package log

import (
	"github.com/sirupsen/logrus"
	"github.com/weekface/mgorus"
	"log"
	"notification/internal/pkg/config"
	"time"
)

type Logger struct {
	cfg *config.Config
}

func NewLogger(cfg *config.Config) *Logger {
	return &Logger{
		cfg: cfg,
	}
}

func (l *Logger) Init() *logrus.Logger {
	logger := logrus.New()
	logger.WithTime(time.Now().In(time.Local))

	mgoHook, err := mgorus.NewHooker(l.cfg.MongoDb.Dsn(), l.cfg.MongoDb.Database, "notification_log")
	if err != nil {
		log.Fatalf("logrus 新增mongo hook失败：%s", err)
	} else {
		logger.AddHook(mgoHook)
	}

	return logger
}
