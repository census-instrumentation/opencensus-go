package tags

type GRPCCodec struct{}

func (c *GRPCCodec) Encode() []byte {}

func (c *GRPCCodec) Decode(b []byte) TagSet {}
