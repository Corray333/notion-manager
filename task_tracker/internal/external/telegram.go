package external

import (
	"log/slog"

	"github.com/Corray333/task_tracker/internal/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ManagerID = 737415136

func (e *External) SendNotification(msg entities.MsgCreator) error {
	message := tgbotapi.NewMessage(ManagerID, msg.ToMsg())
	message.ParseMode = tgbotapi.ModeMarkdown
	_, err := e.tg.GetBot().Send(message)
	if err != nil {
		slog.Error("error sending notification: " + err.Error())
		return err
	}
	return nil
}
