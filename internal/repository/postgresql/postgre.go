/*
Package postgresql: пакет для осуществления взаимодействия с СУБД PostgreSQL. Общение с БД осуществляется через пул
соединений, доступный посредством методов из пакета 'github.com/jackc/pgx'. Методы для взаимодействия с БД содержит
структура PostgreSQL. Функция MustCreate возвращает заполненную структуру PostgreSQL в случае успешной установки связи с
базой данных. В противном случае выполнение приложения прекращается. При отсутствии в базе данных схемы или какой-либо
из необходимых для работы таблиц, они создаются.
*/

package postgresql

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/lazylex/messaggio/internal/config"
	"github.com/lazylex/messaggio/internal/domain/value_objects/status"
	"github.com/lazylex/messaggio/internal/dto"
	"github.com/lazylex/messaggio/internal/ports/repository"
	"log/slog"
	"os"
	"strings"
)

// PostgreSQL структура, хранящая пул соединений, их максимальное количество и текущую схему базы данных.
type PostgreSQL struct {
	pool           *pgx.ConnPool // Пул соединений
	maxConnections int           // Максимально доступное количество соединений с БД
	schema         string        // Схема базы данных
}

// MustCreate возвращает структуру для взаимодействия с базой данных в СУБД PostgreSQL. В случае ошибки завершает
// работу всего приложения.
func MustCreate(cfg config.PersistentStorage) *PostgreSQL {
	schema := "public"
	if len(cfg.DatabaseSchema) > 0 {
		schema = pgx.Identifier{cfg.DatabaseSchema}.Sanitize()
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:          cfg.DatabaseAddress,
			Port:          uint16(cfg.DatabasePort),
			Database:      cfg.DatabaseName,
			User:          cfg.DatabaseLogin,
			Password:      cfg.DatabasePassword,
			RuntimeParams: map[string]string{"search_path": schema},
		},
		MaxConnections: cfg.DatabaseMaxOpenConnections,
	})

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	} else {
		slog.Info("successfully create connection poll to postgres DB")
	}

	client := &PostgreSQL{pool: pool, maxConnections: cfg.DatabaseMaxOpenConnections, schema: schema}

	if err = client.createNotExistedSchemaAndTables(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return client
}

// createNotExistedSchemaAndTables создает схему и таблицы в БД, если они отсутствуют.
func (p *PostgreSQL) createNotExistedSchemaAndTables() error {
	var stmt string

	if len(p.schema) > 0 {
		stmt = `CREATE SCHEMA IF NOT EXISTS ` + p.schema
		if _, err := p.pool.Exec(stmt); err != nil {
			return err
		}
	}

	stmt =
		`DO $$
		BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'msg_status') THEN ` +
			fmt.Sprintf("create type msg_status AS ENUM ('%s', '%s');", status.InProcessing, status.Processed) +
			`END IF;
		END
	$$;`

	if _, err := p.pool.Exec(stmt); err != nil {
		return err
	}

	stmt = `
	CREATE TABLE IF NOT EXISTS messages 
		(	
		    id UUID NOT NULL PRIMARY KEY,
			message bytea NOT NULL,
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    		updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),` +
		fmt.Sprintf("status msg_status NOT NULL DEFAULT '%s')", status.InProcessing)

	if _, err := p.pool.Exec(stmt); err != nil {
		return err
	}

	stmt = `
	CREATE OR REPLACE FUNCTION update_modified_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = now();
		RETURN NEW;
	END;
	$$ language 'plpgsql';
	
	DO
	$$
	BEGIN
		IF NOT EXISTS(SELECT *
							 FROM information_schema.triggers
							 WHERE event_object_table = 'messages'
							 AND trigger_name = 'update_messages_modtime'
							 )
		THEN
			CREATE TRIGGER update_messages_modtime
			BEFORE UPDATE ON messages
			FOR EACH ROW EXECUTE FUNCTION update_modified_column();
		END IF ;
	END;
	$$;`

	if _, err := p.pool.Exec(stmt); err != nil {
		return err
	}

	return nil
}

// SaveMessage сохраняет сообщение и его идентификатор в БД. Статус сообщения сохраняется по умолчанию
// (status.InProcessing).
func (p *PostgreSQL) SaveMessage(ctx context.Context, data dto.MessageID) error {
	stmt := `INSERT INTO messages (id, message) values ($1, $2);`
	_, err := p.pool.ExecEx(ctx, stmt, nil, data.ID, data.Message)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERROR: duplicate key value violates unique constraint") {
			return repository.ErrDuplicateKeyValue
		}

		return err
	}

	return nil
}

// UpdateStatus статус сообщения с идентификатором id обновляется на status.Processed.
func (p *PostgreSQL) UpdateStatus(ctx context.Context, id uuid.UUID) error {
	stmt := `UPDATE messages SET status = $1 WHERE id = $2;`
	_, err := p.pool.ExecEx(ctx, stmt, nil, status.Processed, id)

	return err
}

// ProcessedCount возвращает сумму обработанных сообщений за последний час, день, неделю, месяц.
func (p *PostgreSQL) ProcessedCount(ctx context.Context) (dto.Processed, error) {
	var (
		result dto.Processed
		rows   *pgx.Rows
		err    error
	)

	stmt := `SELECT 
    			COUNT(*) FILTER (WHERE updated_at > NOW() - INTERVAL '1 hours'),
    			COUNT(*) FILTER (WHERE updated_at > CURRENT_DATE - INTERVAL '1 days'),
    			COUNT(*) FILTER (WHERE updated_at > CURRENT_DATE - INTERVAL '1 weeks'),
    			COUNT(*) FILTER (WHERE updated_at > CURRENT_DATE - INTERVAL '1 months')
			FROM messages WHERE status = $1;`

	rows, err = p.pool.QueryEx(ctx, stmt, nil, status.Processed)
	if err != nil {
		return dto.Processed{}, err
	}

	for rows.Next() {
		if err = rows.Scan(&result.InHour, &result.InDay, &result.InWeek, &result.InMonth); err != nil {
			return dto.Processed{}, err
		}
	}

	return result, nil
}
