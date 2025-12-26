package idgen

import "github.com/bwmarrin/snowflake"

type SnowflakeIDGenerator struct {
	node *snowflake.Node
}

func NewSnowflakeClient(machineID int64) (*SnowflakeIDGenerator, error) {
	node, err := snowflake.NewNode(machineID)
	if err != nil {
		return nil, err
	}

	return &SnowflakeIDGenerator{
		node: node,
	}, nil
}

func (s *SnowflakeIDGenerator) Generate() uint64 {
	return uint64(s.node.Generate().Int64())
}
