package status

// StatusDef defines a status code with its display name and terminal flag.
type StatusDef struct {
	Code     string
	Name     string
	Terminal bool
}

// MainItemStatuses is the single source of truth for main item status definitions.
var MainItemStatuses = map[string]StatusDef{
	"pending":     {Code: "pending", Name: "待开始", Terminal: false},
	"progressing": {Code: "progressing", Name: "进行中", Terminal: false},
	"blocking":    {Code: "blocking", Name: "阻塞中", Terminal: false},
	"pausing":     {Code: "pausing", Name: "已暂停", Terminal: false},
	"reviewing":   {Code: "reviewing", Name: "待验收", Terminal: false},
	"completed":   {Code: "completed", Name: "已完成", Terminal: true},
	"closed":      {Code: "closed", Name: "已关闭", Terminal: true},
}

// SubItemStatuses is the single source of truth for sub item status definitions.
var SubItemStatuses = map[string]StatusDef{
	"pending":     {Code: "pending", Name: "待开始", Terminal: false},
	"progressing": {Code: "progressing", Name: "进行中", Terminal: false},
	"blocking":    {Code: "blocking", Name: "阻塞中", Terminal: false},
	"pausing":     {Code: "pausing", Name: "已暂停", Terminal: false},
	"completed":   {Code: "completed", Name: "已完成", Terminal: true},
	"closed":      {Code: "closed", Name: "已关闭", Terminal: true},
}

// GetMainItemStatus returns the StatusDef for a main item status code.
func GetMainItemStatus(code string) (StatusDef, bool) {
	def, ok := MainItemStatuses[code]
	return def, ok
}

// GetSubItemStatus returns the StatusDef for a sub item status code.
func GetSubItemStatus(code string) (StatusDef, bool) {
	def, ok := SubItemStatuses[code]
	return def, ok
}

// IsMainTerminal returns true if the code is a terminal status for main items.
func IsMainTerminal(code string) bool {
	def, ok := MainItemStatuses[code]
	return ok && def.Terminal
}

// IsSubTerminal returns true if the code is a terminal status for sub items.
func IsSubTerminal(code string) bool {
	def, ok := SubItemStatuses[code]
	return ok && def.Terminal
}
