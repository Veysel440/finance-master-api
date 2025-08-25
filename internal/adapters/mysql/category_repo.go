package mysql

import (
	"context"
	"errors"
	"strings"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type CategoryRepo struct{ db *sqlx.DB }

func NewCategoryRepo(db *sqlx.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) List(userID int64, typ string) ([]ports.Category, error) {
	var rows []ports.Category
	q := `SELECT id, user_id, name, type FROM categories WHERE user_id=?`
	args := []any{userID}
	if s := strings.TrimSpace(typ); s != "" {
		q += ` AND type=?`
		args = append(args, s)
	}
	q += ` ORDER BY type, name`
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *CategoryRepo) Create(userID int64, c *ports.Category) error {
	res, err := r.db.ExecContext(context.Background(),
		`INSERT INTO categories(user_id, name, type) VALUES (?,?,?)`,
		userID, c.Name, c.Type,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = id
	return nil
}

func (r *CategoryRepo) Update(userID int64, c *ports.Category) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE categories SET name=?, type=? WHERE id=? AND user_id=?`,
		c.Name, c.Type, c.ID, userID,
	)
	return err
}

func (r *CategoryRepo) Delete(userID int64, id int64) error {
	_, err := r.db.ExecContext(context.Background(),
		`DELETE FROM categories WHERE id=? AND user_id=?`, id, userID,
	) // FK RESTRICT -> ilişkili tx varsa 1451
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1451 {
			return errs.HasTransactions
		}
		return err
	}
	return nil
}
