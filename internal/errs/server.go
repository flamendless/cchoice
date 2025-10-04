package errs

import "errors"

var (
	ErrServerInit                 = errors.New("[Server]: Failed to initialize server")
	ErrServerUnimplementedGateway = errors.New("[Server]: Unimplemented payment gateway")
	ErrServerNoMapsFound          = errors.New("[Server]: No maps found")
	ErrServerFSNotSetup           = errors.New("[Server]: Filesystem not setup")
)
