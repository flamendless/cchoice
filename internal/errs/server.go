package errs

import "errors"

var (
	ErrServerInit                 = errors.New("[SERVER]: Failed to initialize server")
	ErrServerUnimplementedGateway = errors.New("[SERVER]: Unimplemented gateway")
	ErrServerNoMapsFound          = errors.New("[SERVER]: No maps found")
	ErrServerFSNotSetup           = errors.New("[SERVER]: Filesystem not setup")
)
