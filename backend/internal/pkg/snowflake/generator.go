package snowflake

import "github.com/bwmarrin/snowflake"

var node *snowflake.Node

// Init initializes the snowflake node with the given workerID (0-1023).
// Must be called once at application startup before Generate is used.
func Init(workerID int64) error {
	var err error
	node, err = snowflake.NewNode(workerID)
	return err
}

// Generate returns a new unique int64 biz_key.
func Generate() int64 {
	return node.Generate().Int64()
}
