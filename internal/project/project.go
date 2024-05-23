package project

import (
	"context"

	"github.com/jomei/notionapi"
)

type Project struct {
	ProjectID       string `db:"project_id"`
	Name            string `db:"name"`
	TimeDBID        string `db:"time_db_id"`
	TasksDBID       string `db:"tasks_db_id"`
	WorkersDBID     string `db:"workers_db_id"`
	TasksLastSynced int64  `db:"tasks_last_synced"`
	TimeLastSynced  int64  `db:"time_last_synced"`
}

func (p *Project) Update(client notionapi.Client) error {
	_, err := client.Database.Get(context.Background(), notionapi.DatabaseID(p.TimeDBID))
	if err != nil {
		return err
	}
	client.Search.Do(context.Background(), &notionapi.SearchRequest{
		Filter: notionapi.SearchFilter{
			Property: "database_id",
			Value:    p.TasksDBID,
		},
	})
	return nil
}
