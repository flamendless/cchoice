package auth

import (
	"context"
)

type ClientToken struct {
	Token string
	TSC   bool
}

func (t ClientToken) GetRequestMetadata(
	ctx context.Context,
	in ...string,
) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.Token,
	}, nil
}

func (t ClientToken) RequireTransportSecurity() bool {
	return t.TSC
}
