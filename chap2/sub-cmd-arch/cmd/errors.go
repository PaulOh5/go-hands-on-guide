package cmd

import "errors"

var ErrorNoServerSpecified = errors.New("you have to specify the remote server")

var ErrorInvalidHttpMethod = errors.New("invalid HTTP method")
