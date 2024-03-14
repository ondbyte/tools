package parser

import (
	"fmt"
	"go/token"
	"testing"
)

// @handler("GET","/users/{userId}/orders/{orderId}")
// @description("xyz")
func HandleA(
	// @pathparam("userId")
	// @description("orderId description")
	userId string,
	// @pathparam("userId")
	// @description("orderId description")
	orderId uint,
) {

}

var source = `package yadu
// @handler("GET","/users/{userId}/orders/{orderId}")
// @description("xyz")
func HandleA(
	// @pathparam("userId")
	// @description("orderId description")
	userId string,
	// @pathparam("userId")
	// @description("orderId description")
	orderId uint,
) {

}`

func TestDecors(t *testing.T) {
	df, f, err := ParseFile(token.NewFileSet(), "yadu.go", source, ParseComments)
	fmt.Println(df,f,err)
}
