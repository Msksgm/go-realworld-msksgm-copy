package postgres

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/msksgm/go-realworld-msksgm-copy/conduit"
)

var _ conduit.ArticleService = (*ArticleService)(nil)

type ArticleService struct {
	db *DB
}

func NewArticleService(db *DB) *ArticleService {
	return &ArticleService{db}
}

func (as *ArticleService) CreateArticle(ctx context.Context, article *conduit.Article) error {
	tx, err := as.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createArticle(ctx, tx, article); err != nil {
		return err
	}

	return tx.Commit()
}

func createArticle(ctx context.Context, tx *sqlx.Tx, article *conduit.Article) error {
	query := `
	INSERT INTO articles (title, body, description, author_id, slug) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at
	`

	args := []interface{}{
		article.Title,
		article.Body,
		article.Description,
		article.Author.ID,
		article.Slug,
	}

	err := tx.QueryRowxContext(ctx, query, args...).Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)
	if err != nil {
		return err
	}

	tags := make([]string, len(article.Tags))
	for i, tag := range article.Tags {
		tags[i] = tag.Name
	}

	err = setArticleTags(ctx, tx, article, tags)
	if err != nil {
		return err
	}

	return nil
}

func setArticleTags(ctx context.Context, tx *sqlx.Tx, article *conduit.Article, tags []string) error {
	for _, v := range tags {
		tag, err := findTagByName(ctx, tx, v)
		if err != nil {
			switch {
			case errors.Is(err, conduit.ErrNotFound):
				tag = &conduit.Tag{Name: v}
				err = createTag(ctx, tx, tag)
				if err != nil {
					return err
				}
			default:
				return err
			}
		}

		err = associateArticleWithTag(ctx, tx, article, tag)

		if err != nil {
			return err
		}
	}

	return nil
}

func associateArticleWithTag(ctx context.Context, tx *sqlx.Tx, article *conduit.Article, tag *conduit.Tag) error {
	query := "INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2)"

	_, err := tx.ExecContext(ctx, query, article.ID, tag.ID)
	if err != nil {
		return err
	}

	return nil
}
