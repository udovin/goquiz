package migrations

import (
	"context"

	"github.com/udovin/gosql"
	"github.com/udovin/solve/db"
	"github.com/udovin/solve/db/schema"
)

func init() {
	db.RegisterMigration(&m001{})
}

type m001 struct{}

func (m *m001) Name() string {
	return "001_initial"
}

func (m *m001) Apply(ctx context.Context, conn *gosql.DB) error {
	tx := db.GetRunner(ctx, conn)
	for _, table := range m001Tables {
		query, err := table.BuildCreateSQL(conn.Dialect(), false)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (m *m001) Unapply(ctx context.Context, conn *gosql.DB) error {
	tx := db.GetRunner(ctx, conn)
	for i := 0; i < len(m001Tables); i++ {
		table := m001Tables[len(m001Tables)-i-1]
		query, err := table.BuildDropSQL(conn.Dialect(), false)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

var m001Tables = []schema.Table{
	{
		Name: "goquiz_setting",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "key", Type: schema.String},
			{Name: "value", Type: schema.String},
		},
	},
	{
		Name: "goquiz_setting_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "key", Type: schema.String},
			{Name: "value", Type: schema.String},
		},
	},
	{
		Name: "goquiz_task",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "status", Type: schema.Int64},
			{Name: "kind", Type: schema.Int64},
			{Name: "config", Type: schema.JSON},
			{Name: "state", Type: schema.JSON},
			{Name: "expire_time", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_task_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "status", Type: schema.Int64},
			{Name: "kind", Type: schema.Int64},
			{Name: "config", Type: schema.JSON},
			{Name: "state", Type: schema.JSON},
			{Name: "expire_time", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_role",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: schema.String},
		},
	},
	{
		Name: "goquiz_role_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "name", Type: schema.String},
		},
	},
	{
		Name: "goquiz_role_edge",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "role_id", Type: schema.Int64},
			{Name: "child_id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_role_edge_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "role_id", Type: schema.Int64},
			{Name: "child_id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_account",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "kind", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_account_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "kind", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_account_role",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "account_id", Type: schema.Int64},
			{Name: "role_id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_account_role_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "account_id", Type: schema.Int64},
			{Name: "role_id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_session",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "account_id", Type: schema.Int64},
			{Name: "secret", Type: schema.String},
			{Name: "create_time", Type: schema.Int64},
			{Name: "expire_time", Type: schema.Int64},
			{Name: "remote_addr", Type: schema.String},
			{Name: "user_agent", Type: schema.String},
		},
	},
	{
		Name: "goquiz_session_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "account_id", Type: schema.Int64},
			{Name: "secret", Type: schema.String},
			{Name: "create_time", Type: schema.Int64},
			{Name: "expire_time", Type: schema.Int64},
			{Name: "remote_addr", Type: schema.String},
			{Name: "user_agent", Type: schema.String},
		},
	},
	{
		Name: "goquiz_user",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "account_id", Type: schema.Int64},
			{Name: "login", Type: schema.String},
			{Name: "password_hash", Type: schema.String},
			{Name: "password_salt", Type: schema.String},
			{Name: "email", Type: schema.String, Nullable: true},
			{Name: "first_name", Type: schema.String, Nullable: true},
			{Name: "last_name", Type: schema.String, Nullable: true},
			{Name: "middle_name", Type: schema.String, Nullable: true},
		},
	},
	{
		Name: "goquiz_user_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
			{Name: "account_id", Type: schema.Int64},
			{Name: "login", Type: schema.String},
			{Name: "password_hash", Type: schema.String},
			{Name: "password_salt", Type: schema.String},
			{Name: "email", Type: schema.String, Nullable: true},
			{Name: "first_name", Type: schema.String, Nullable: true},
			{Name: "last_name", Type: schema.String, Nullable: true},
			{Name: "middle_name", Type: schema.String, Nullable: true},
		},
	},
	{
		Name: "goquiz_visit",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "time", Type: schema.Int64},
			{Name: "account_id", Type: schema.Int64, Nullable: true},
			{Name: "session_id", Type: schema.Int64, Nullable: true},
			{Name: "host", Type: schema.String},
			{Name: "protocol", Type: schema.String},
			{Name: "method", Type: schema.String},
			{Name: "remote_addr", Type: schema.String},
			{Name: "user_agent", Type: schema.String},
			{Name: "path", Type: schema.String},
			{Name: "real_ip", Type: schema.String},
			{Name: "status", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_quiz",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
		},
	},
	{
		Name: "goquiz_quiz_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_pool",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
		},
	},
	{
		Name: "goquiz_pool_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
		},
	},
	{
		Name: "goquiz_problem",
		Columns: []schema.Column{
			{Name: "id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
		},
	},
	{
		Name: "goquiz_problem_event",
		Columns: []schema.Column{
			{Name: "event_id", Type: schema.Int64, PrimaryKey: true, AutoIncrement: true},
			{Name: "event_kind", Type: schema.Int64},
			{Name: "event_time", Type: schema.Int64},
			{Name: "event_account_id", Type: schema.Int64, Nullable: true},
			{Name: "id", Type: schema.Int64},
		},
	},
}
