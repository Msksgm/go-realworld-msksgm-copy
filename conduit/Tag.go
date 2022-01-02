package conduit

type Tag struct {
	ID   uint
	Name string
}

type TagFilter struct {
	Name *string

	Limit  int
	Offset int
}
