package adapter

import (
	"fmt"

	"github.com/google/instrumentation-go/stats/tagging"
)

func encodeToGrpcFormat(ts tagging.TagsSet) ([]byte, error) {
	return nil, nil
}

func decodeFromGrpcFormat(bytes []byte) (tagging.TagsSet, error) {
	var err error
	return nil, fmt.Errorf("decodeFromGrpcFormat failed to construct tagging.TagsSet from []byte: %v. %v", bytes, err)
}
