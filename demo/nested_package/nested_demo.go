package nested_package

import "github.com/difof/errors"

func CreateNestedError(chain error, depth int) error {
	return errors.Wrapf(chain, "nested error at depth %d", depth)
}
