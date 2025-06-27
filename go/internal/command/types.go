/*
 * JsonDB Go 版本 - 命令類型定義
 *
 * 定義所有支援的命令類型和結構
 */

package command

type CommandType int

const (
	SELECT CommandType = iota

	// * KV 操作
	GET
	SET
	DEL
	EXISTS
	KEYS
	TYPE

	// * DOC 操作
	FIND
	SORT
	ADD
	UPDATE
	REMOVE

	// * TTL 操作
	TTL
	EXPIRE
	EXPIREAT
	PERSIST

	// * 其他操作
	HELP
	PING
)

type Command struct {
	Type CommandType
	Args map[string]interface{}
}

func NewCommand(cmd CommandType) *Command {
	return &Command{
		Type: cmd,
		Args: make(map[string]interface{}),
	}
}

func (c *Command) SetArg(key string, value interface{}) {
	c.Args[key] = value
}

func (c *Command) GetArg(key string) (interface{}, bool) {
	value, isExist := c.Args[key]
	return value, isExist
}

func (c *Command) GetStr(key string) string {
	if value, isExist := c.Args[key]; isExist {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *Command) GetStrAry(key string) []string {
	if value, isExist := c.Args[key]; isExist {
		if slice, ok := value.([]string); ok {
			return slice
		}
	}
	return nil
}

func (c *Command) GetInt(key string) int {
	if value, isExist := c.Args[key]; isExist {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

func (c *Command) GetUint64(key string) uint64 {
	if value, isExist := c.Args[key]; isExist {
		if u, ok := value.(uint64); ok {
			return u
		}
	}
	return 0
}
