package template

const (
	// Imports defines a import template for model in cache case
	Imports = `import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	{{if .time}}"time"{{end}}
	"comm/util"

	proto "proto/{{.serviceName}}"
	"github.com/Masterminds/squirrel"
	sqlBuilder "github.com/lann/builder"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)
`
	// ImportsNoCache defines a import template for model in normal case
	ImportsNoCache = `import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	{{if .time}}"time"{{end}}
	"comm/util"
	
	proto "proto/{{.serviceName}}"
	"github.com/Masterminds/squirrel"
	sqlBuilder "github.com/lann/builder"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)
`
)
