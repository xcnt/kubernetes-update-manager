package gocheck_matchers

import (
	. "gopkg.in/check.v1"
)

type isFalseChecker struct {
	*CheckerInfo
}

func (checker *isFalseChecker) Check(params []interface{}, names []string) (result bool, error string) {
	switch v := params[0].(type) {
	case bool:
		result = !v
	default:
		error = "Not a boolean value provided"
	}
	return
}

var IsFalse = &isFalseChecker{
	CheckerInfo: &CheckerInfo{Name: "Checks whether the entry is false", Params: []string{"value"}},
}

