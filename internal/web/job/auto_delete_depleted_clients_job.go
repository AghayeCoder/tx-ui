package job

import (
	"time"

	"x-ui/internal/database"
	"x-ui/internal/database/model"
	"x-ui/internal/web/service"

	"github.com/robfig/cron/v3"
)

var autoDeleteDepletedClientsJob *AutoDeleteDepletedClientsJob

type AutoDeleteDepletedClientsJob struct {
	cron     *cron.Cron
	settingService *service.SettingService
	inboundService *service.InboundService
}

func NewAutoDeleteDepletedClientsJob() *AutoDeleteDepletedClientsJob {
	return &AutoDeleteDepletedClientsJob{
		cron:     cron.New(),
		settingService: &service.SettingService{},
		inboundService: &service.InboundService{},
	}
}

func (j *AutoDeleteDepletedClientsJob) Start() {
	_, err := j.cron.AddFunc("@daily", j.Run)
	if err != nil {
		panic(err)
	}
	j.cron.Start()
}

func (j *AutoDeleteDepletedClientsJob) Stop() {
	j.cron.Stop()
}

type ClientWithInboundId struct {
	model.Client
	InboundId int `json:"inbound_id"`
}

func (j *AutoDeleteDepletedClientsJob) Run() {
	autoDeleteDay, err := j.settingService.GetAutoDeleteDay()
	if err != nil {
		return
	}
	if autoDeleteDay == 0 {
		return
	}

	db := database.GetDB()
	clients := make([]*ClientWithInboundId, 0)
	expiryLimit := time.Now().Unix()*1000 - int64(autoDeleteDay)*24*60*60*1000
	db.Model(&model.Client{}).
		Select("clients.*, client_traffics.inbound_id").
		Joins("join client_traffics on clients.email = client_traffics.email").
		Where("clients.enable = ?", true).
		Where("(client_traffics.total_gb > 0 AND (client_traffics.up_gb + client_traffics.down_gb) >= client_traffics.total_gb) OR (clients.expiry_time != 0 AND clients.expiry_time < ?)", expiryLimit).
		Find(&clients)

	for _, client := range clients {
		j.inboundService.DelInboundClient(client.InboundId, client.Email)
	}
}
