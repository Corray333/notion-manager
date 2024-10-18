package external

import (
	"fmt"
	"strings"

	"github.com/Corray333/task_tracker/internal/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// const ManagerID = 737415136
const ManagerID = 377742748

func (e *External) SendNotification(rows []entities.Row) error {
	msgText := fmt.Sprintf("Ошибки в записях пользователя %s:\n", rows[0].Employee)

	for i, row := range rows {
		msgText += fmt.Sprintf("%d. <a href=\"notion.so/%s\">%s</a>\n", i+1, strings.ReplaceAll(row.ID, "-", ""), row.Description)
	}

	msg := tgbotapi.NewMessage(ManagerID, msgText)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := e.tg.GetBot().Send(msg); err != nil {
		return err
	}
	return nil
}
