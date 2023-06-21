package errs

import "errors"

var ErrInvalidPublishedAt = errors.New("cannot parse publishedAt")
