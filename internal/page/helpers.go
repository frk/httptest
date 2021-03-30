package page

import (
	"html/template"
	"strings"
)

var helpers = template.FuncMap{
	"lower": strings.ToLower,
}
