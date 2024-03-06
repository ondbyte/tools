package parser

import (
	"fmt"
	"go/ast"
	"strconv"
)

func VerifyPathParamDecor(decorName string, args []ast.Expr, paramName string, paramType ast.Expr) (*PathParamDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(decorName, string(PATH), args, 1)
	if err != nil {
		return nil, err
	}
	if paramValues == nil {
		return nil, nil
	}
	return &PathParamDecor{
		PathParamName: paramValues[0],
	}, nil
}

var allowedHttpMethods = []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"}

func VerifyHandlerDecor(decorName string, args []ast.Expr) (*HandlerDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(decorName, string(HANDLER), args, 2)
	if err != nil {
		return nil, err
	}
	if paramValues == nil {
		return nil, nil
	}
	_method := ""
	for _, method := range allowedHttpMethods {
		if paramValues[0] == method {
			_method = method
			break
		}
	}
	if _method == "" {
		return nil, &DecorationErr{
			pos: args[0].End(),
			msg: fmt.Sprintf("first argument should be string http method, allowed values are %v", allowedHttpMethods),
		}
	}
	PathParams := map[string]bool{}
	allMatches := endpointContentRegex.FindAllStringSubmatch(paramValues[1], -1)
	for _, match := range allMatches {
		paramValues = append(paramValues, match[1])
	}
	return &HandlerDecor{
		HttpMethod: _method,
		Path:       paramValues[1],
		PathParams: PathParams,
	}, nil
}

func VerifyDescrDecor(decorName string, args []ast.Expr) (*DescriptionDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(decorName, string(DESCR), args, 1)
	if err != nil {
		return nil, err
	}
	if paramValues == nil {
		return nil, nil
	}
	return &DescriptionDecor{
		Data: paramValues[0],
	}, nil
}

// verifies a decor name and arguments it must take
// returns the argument values or err
func VerifyDecorArgs(decorName, requiredName string, args []ast.Expr, noOfRequiredParam int) (r []string, d *DecorationErr) {
	if decorName != requiredName {
		return nil, nil
	}
	if len(args) != noOfRequiredParam {
		return nil, &DecorationErr{
			pos: args[len(args)-1].End(),
			msg: fmt.Sprintf("%v requires %v arguments but you have provided %v", requiredName, noOfRequiredParam, len(args)),
		}
	}
	r = []string{}
	for _, a := range args {
		lit, ok := a.(*ast.BasicLit)
		var litVal string
		if !ok {
			return nil, &DecorationErr{
				pos: a.Pos(),
				msg: fmt.Sprintf("only string param values are supported for now"),
			}
		}
		litVal, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, &DecorationErr{
				pos: a.Pos(),
				msg: fmt.Sprintf("only string param values are supported for now"),
			}
		}
		r = append(r, litVal)
	}
	return
}
