package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/msksgm/go-realworld-msksgm-copy/conduit"
)

func createTag(ctx context.Context, tx *sqlx.Tx, tag *conduit.Tag) error {
	query := "INSERT INTO tags (name) VALUES ($1) RETURNING id"

	err := tx.QueryRowxContext(ctx, query, tag.Name).Scan(&tag.ID)
	if err != nil {
		return err
	}

	return nil
}

func findTags(ctx context.Context, tx *sqlx.Tx, filter conduit.TagFilter) ([]*conduit.Tag, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0

	if v := filter.Name; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("name = $%d", argPosition)), append(args, *v)
	}

	query := "SELECT * from tags " + formatWhereClause(where) + " ORDER BY id ASC"
	tags := make([]*conduit.Tag, 0)
	err := findMany(ctx, tx, &tags, query, args...)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func findTagByName(ctx context.Context, tx *sqlx.Tx, name string) (*conduit.Tag, error) {
	ts, err := findTags(ctx, tx, conduit.TagFilter{Name: &name})
	if err != nil {
		return nil, err
	} else if len(ts) == 0 {
		return nil, conduit.ErrNotFound
	}

	return ts[0], nil
}
