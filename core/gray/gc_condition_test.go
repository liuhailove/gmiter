package gray

import (
	"fmt"
	"testing"
)

func TestConditionSelect(t *testing.T) {
	var res = "ABc"
	if res[len(res)-1] == '*' {
		res = "hello"
	}
	fmt.Println(res)
}
