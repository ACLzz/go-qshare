package conn

import "errors"

var ErrIntroductionCantBeEmpty = errors.New("introduction must contain " +
	"at least one file or text payload")
