package x

import (
	"fmt"
	"testing"
)

func TestMD(t *testing.T) {
	code := CodeLines(
		`// handler("id")`,
		`func YourFunc(`,
		`	// path("id")`,
		`	id string,`,
		`){`,
		``,
		`}`,
	)
	fmt.Println(code)
}

func CodeLines(lines ...string) (c string) {
	c = "```go\n"
	for _, l := range lines {
		c += l + "\n"
	}
	c += "```"
	return
}
