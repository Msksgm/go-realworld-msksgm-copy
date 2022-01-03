package postgres

import (
	"context"
	"errors"
	"fmt"

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

func (as *ArticleService) Articles(ctx context.Context, filter conduit.ArticleFilter) ([]*conduit.Article, error) {
	tx, err := as.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	articles, err := findArticles(ctx, tx, filter)
	if err != nil {
		return nil, err
	}

	return articles, tx.Commit()
}

func (as *ArticleService) ArticleFeed(ctx context.Context, user *conduit.User, filter conduit.ArticleFilter) ([]*conduit.Article, error) {
	tx, err := as.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	articles, err := getArticlesFromUserFollowings(ctx, tx, user, filter)
	if err != nil {
		return nil, err
	}

	return articles, tx.Commit()
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

func findArticles(ctx context.Context, tx *sqlx.Tx, filter conduit.ArticleFilter) ([]*conduit.Article, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0 // used to set correct postgres argument enums i.e $1, $2

	if v := filter.ID; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("id = $%d", argPosition)), append(args, *v)
	}

	if v := filter.AuthorID; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("author_id = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Slug; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("slug = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Title; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("title = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Tag; v != nil {
		argPosition++
		clause := `id IN (select article_id from article_tags where tag_id in (
			   select id from tags where name = $%d)
		    )`
		where, args = append(where, fmt.Sprintf(clause, argPosition)), append(args, *v)
	}

	if v := filter.AuthorUsername; v != nil {
		argPosition++
		clause := "author_id = (select id from users where username = $%d)"
		where, args = append(where, fmt.Sprintf(clause, argPosition)), append(args, *v)
	}

	if v := filter.FavoritedBy; v != nil {
		argPosition++
		clause := `id IN (select article_id from favorites where user_id = (
			select id from users where username = $%d LIMIT 1)
			)`
		where, args = append(where, fmt.Sprintf(clause, argPosition)), append(args, *v)
	}

	query := "SELECT * from articles" + formatWhereClause(where) + " ORDER BY created_at DESC"
	articles, err := queryArticles(ctx, tx, query, args...)
	if err != nil {
		return articles, err
	}

	return articles, nil
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

func attachArticleAssociations(ctx context.Context, tx *sqlx.Tx, article *conduit.Article) error {
	tags, err := findArticleTags(ctx, tx, article)
	if err != nil {
		return fmt.Errorf("cannot find article tags: %w", err)
	}

	article.Tags = tags

	user, err := findUserByID(ctx, tx, article.AuthorID)
	if err != nil {
		return fmt.Errorf("cannot find article author: %w", err)
	}

	article.Author = user

	query := `SELECT * from users WHERE id IN (
		SELECT user_id FROM favorites WHERE article_id = $1
	)`

	favorites := make([]*conduit.User, 0)

	if err := findMany(ctx, tx, &favorites, query, article.ID); err != nil {
		return err
	}

	article.FavoritedBy = favorites
	article.FavoritesCount = int64(len(favorites))

	return nil
}

func findArticleTags(ctx context.Context, tx *sqlx.Tx, article *conduit.Article) ([]*conduit.Tag, error) {
	query := `
	SELECT * from tags WHERE id IN (
		SELECT tag_id FROM article_tags WHERE article_id = $1
	)
	`
	tags := make([]*conduit.Tag, 0)
	if err := findMany(ctx, tx, &tags, query, article.ID); err != nil {
		return tags, err
	}
	return tags, nil
}

func getArticlesFromUserFollowings(ctx context.Context, tx *sqlx.Tx, user *conduit.User, filter conduit.ArticleFilter) ([]*conduit.Article, error) {
	query := `
	SELECT * from articles as a WHERE author_id IN (
		SELECT following_id from followings WHERE follower_id = $1 
	) ORDER BY a.created_at DESC
	` + formatLimitOffset(filter.Limit, filter.Offset)

	return queryArticles(ctx, tx, query, user.ID)
}

func queryArticles(ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]*conduit.Article, error) {
	articles := make([]*conduit.Article, 0)
	err := findMany(ctx, tx, &articles, query, args...)
	if err != nil {
		return articles, err
	}

	for _, article := range articles {
		if err := attachArticleAssociations(ctx, tx, article); err != nil {
			return nil, err
		}
	}

	return articles, nil
}
