package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

const PATH decoratorName = "path"

func (dc *DecorComment) VerifyPathParamDecor() (*PathParamDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(dc.DecorName, string(PATH), dc.Args, token.STRING)
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

const HANDLER decoratorName = "handler"

func (dc *DecorComment) VerifyHandlerDecor() (*HandlerDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(dc.DecorName, string(HANDLER), dc.Args, token.STRING, token.STRING)
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
			pos: dc.Args[0].End(),
			msg: fmt.Sprintf("first argument should be string http method, allowed values are %v", allowedHttpMethods),
		}
	}
	PathParams := map[string]bool{}
	allMatches := endpointContentRegex.FindAllStringSubmatch(paramValues[1], -1)
	for _, match := range allMatches {
		PathParams[match[1]] = true
	}
	return &HandlerDecor{
		HttpMethod: _method,
		Path:       paramValues[1],
		PathParams: PathParams,
	}, nil
}

const DESCR decoratorName = "description"

func (dc *DecorComment) VerifyDescrDecor() (*DescriptionDecor, *DecorationErr) {
	paramValues, err := VerifyDecorArgs(dc.DecorName, string(DESCR), dc.Args, token.STRING)
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
func VerifyDecorArgs(decorName *ast.Ident, reqDecorName string, args []*ast.BasicLit, requiredArgs ...token.Token) (r []string, d *DecorationErr) {
	if decorName.Name != reqDecorName {
		return nil, nil
	}
	if len(args) != len(requiredArgs) {
		if len(args) == 0 {
			return nil, &DecorationErr{
				pos: decorName.End() + 1,
				msg: fmt.Sprintf("%v requires %v arguments but you have provided %v", reqDecorName, len(requiredArgs), len(args)),
			}
		}
		return nil, &DecorationErr{
			pos: args[len(args)-1].End(),
			msg: fmt.Sprintf("%v requires %v arguments but you have provided %v", reqDecorName, len(requiredArgs), len(args)),
		}
	}
	r = []string{}
	for i, a := range args {
		b := requiredArgs[i]
		if b != a.Kind {
			return nil, &DecorationErr{
				pos: a.Pos(),
				msg: fmt.Sprintf("expected param %v to be a %v", i, b.String()),
			}
		}
		us, err := strconv.Unquote(a.Value)
		if err != nil {
			return nil, &DecorationErr{
				pos: a.Pos(),
				msg: err.Error(),
			}
		}
		r = append(r, us)
	}
	return
}
