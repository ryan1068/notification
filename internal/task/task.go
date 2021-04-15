package task

import (
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"notification/internal/entity"
	"notification/internal/notify"
	"notification/internal/pkg/config"
	"notification/internal/target"
	"sync"
	"time"
)

type task struct {
	db      *gorm.DB
	mongodb *mgo.Database
	logger  *logrus.Logger
	cfg     *config.Config
	mu      sync.RWMutex
}

func NewTask(db *gorm.DB, mongodb *mgo.Database, logger *logrus.Logger, cfg *config.Config) *task {
	return &task{
		db:      db,
		mongodb: mongodb,
		logger:  logger,
		cfg:     cfg,
	}
}

// 发送系统消息通知
func (t *task) SendNotification(doneChan chan int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	f := func(doneChan chan int) ([]entity.Notify, error) {
		notifyRepository := notify.NewRepository(t.db, t.logger)
		notifies, err := notifyRepository.FindAll("type = 0 and send_type = 0 and send_status = 0 and send_time <= ?", time.Now().Unix())
		if err != nil {
			return nil, err
		}
		if len(notifies) > 0 {
			targetRepository := target.NewRepository(t.db, t.logger)
			for _, notify := range notifies {
				go func(notify entity.Notify) error {
					admins, err := t.getGroupAdmins(notify)
					if err != nil {
						t.logger.WithFields(logrus.Fields{
							"notify_id": notify.ID,
						}).Error(err.Error())
						return err
					}
					for _, admin := range admins {
						_, _ = targetRepository.Create(&entity.NotifyTarget{
							NotifyID:  notify.ID,
							Scene:     0,
							AdminID:   admin.ID,
							CreatedAt: time.Now().Unix(),
						})
					}

					doneChan <- notify.ID
					return nil
				}(notify)
			}
		}
		return notifies, nil
	}

	return t.log(doneChan, f)
}

func (t *task) log(doneChan chan int, f func(doneChan chan int) ([]entity.Notify, error)) error {
	id := bson.NewObjectId()
	collection := t.mongodb.C("notification_task")
	err := collection.Insert(map[string]interface{}{
		"_id":        id,
		"start_time": time.Now().Unix(),
		"date":       time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		"status":     0,
	})
	if err != nil {
		return nil
	}

	res, err := f(doneChan)
	if err != nil {
		err = collection.UpdateId(id,
			bson.M{"$set": bson.M{
				"status":   2,
				"end_time": time.Now().Unix(),
				"err":      err.Error(),
				"res_msg":  "执行失败",
			}})
		if err != nil {
			return nil
		}
	} else {
		err = collection.UpdateId(id,
			bson.M{"$set": bson.M{
				"status":   1,
				"end_time": time.Now().Unix(),
				"data":     res,
				"res_msg":  "执行成功",
			}})
		if err != nil {
			return nil
		}
	}
	return nil
}

// 发送完成，更新任务状态
func (t *task) Done(id int) (*entity.Notify, error) {
	notifyRepository := notify.NewRepository(t.db, t.logger)
	return notifyRepository.Update(id, &entity.Notify{SendStatus: 1, CompletionTime: time.Now().Unix()})
}

// 获取通知的集团管理员
func (t *task) getGroupAdmins(notify entity.Notify) ([]entity.Admin, error) {
	group, err := notify.GetGroup(t.db)
	if err != nil {
		return nil, err
	}
	return group.GetAdmins(t.db)
}
