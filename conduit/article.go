package conduit

import (
	"context"
	"time"
)

type Article struct {
	ID             uint      `json:"-"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	Description    string    `json:"description"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int64     `json:"favoritesCount" db:"favorites_count"`
	FavoritedBy    []*User   `json:"-"`
	Slug           string    `json:"slug"`
	AuthorID       uint      `json:"-" db:"author_id"`
	Author         *User     `json:"-"`
	AuthorProfile  *Profile  `json:"author"`
	Tags           []*Tag    `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type ArticleFilter struct {
	ID             *uint
	Title          *string
	Description    *string
	AuthorID       *uint
	AuthorUsername *string
	Tag            *string
	Slug           *string
	FavoritedBy    *string

	Limit  int
	Offset int
}

type ArticleService interface {
	CreateArticle(context.Context, *Article) error
}

func (a *Article) AddTags(_tags ...string) {
	for _, t := range _tags {
		a.Tags = append(a.Tags, &Tag{Name: t})
	}
}
