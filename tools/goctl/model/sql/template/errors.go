package template

// Error defines an error template
const Error = `package {{.pkg}}

import (
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)


var (
	ErrNotFound = sqlx.ErrNotFound
	errNotSelectBuilder = errors.New("not SelectBuilder")
)
`
