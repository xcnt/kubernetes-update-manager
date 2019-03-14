package gocheck_matchers

import (
	. "gopkg.in/check.v1"
)

type isTrueChecker struct {
	*CheckerInfo
}

func (checker *isTrueChecker) Check(params []interface{}, names []string) (result bool, error string) {
	switch v := params[0].(type) {
	case bool:
		result = v
	default:
		error = "Not a boolean value provided"
	}
	return
}

var IsTrue = &isTrueChecker{
	CheckerInfo: &CheckerInfo{Name: "Checks whether the entry is true", Params: []string{"value"}},
}

