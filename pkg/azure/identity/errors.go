package identity

import "errors"

var (
	ErrInvalidResponseCode           = errors.New("invalid response code")
	ErrAuthorizationEndpointNotFound = errors.New("authorize endpoint wasn't found")
	ErrTokenEndpointNotFound         = errors.New("token endpoint not found")
	ErrIssuerNotFound                = errors.New("issues not found")
	ErrUnknownAccountType            = errors.New("unknown account type")
)
