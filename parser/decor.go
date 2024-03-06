package parser

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"
	"go/token"
)

type decoratorName string

var (
	PATH    decoratorName = "path"
	DESCR   decoratorName = "description"
	HANDLER decoratorName = "handler"
)

type Decorator interface {
	decoratorName() decoratorName
}

// a file which contains possible decorators for multiple Decl's
type DecoratedFile struct {
	decorations map[ast.Decl]*DeclDecorators
}

// a Decl which has decorators
type DecoratedDecl interface {
	decoratedDecl()
}

type DecoratedFuncDecl struct {
}

// decoratedDecl implements DecoratedDecl.
func (d *DecoratedFuncDecl) decoratedDecl() {
	panic("unimplemented")
}

var _ DecoratedDecl = &DecoratedFuncDecl{}

var endpointRegex = regexp.MustCompile(string(HANDLER) + `\("([^"]+)"\)`)
var endpointContentRegex = regexp.MustCompile(`{([^}]*)}`)

type PathParamDecor struct {
	Pos           token.Pos
	PathParamName string
}

// decorator implements Decorator.
func (p *PathParamDecor) decoratorName() decoratorName {
	return PATH
}

var pathRegex = regexp.MustCompile(`path\("([^"]+)"\)`)

type Query struct {
	Pos    token.Pos
	Params map[string]string
}

var queryRegex = regexp.MustCompile(`query\("([^"]+)"\)`)

type Header struct {
	Pos    token.Pos
	Params map[string]string
}

var headerRegex = regexp.MustCompile(`header\("([^"]+)"\)`)

type Body struct {
	Pos    token.Pos
	Params map[string]string
}

var bodyRegex = regexp.MustCompile(`body\("([^"]+)"\)`)

type DescriptionDecor struct {
	Pos  token.Pos
	Data string
}

// decorator implements Decorator.
func (d *DescriptionDecor) decoratorName() decoratorName {
	return DESCR
}

type HandlerDecor struct {
	HttpMethod string
	Path       string
	PathParams map[string]bool
}

// decorator implements Decorator.
func (h *HandlerDecor) decoratorName() decoratorName {
	return HANDLER
}

// handlerDecor implements HandlerDecorExpr.
func (d *DescriptionDecor) handlerDecor() {
	panic("unimplemented")
}

var descriptionRegex = regexp.MustCompile(`description\("([^"]+)"\)`)

type DecorationErr struct {
	pos token.Pos
	msg string
}

func (d *DecorationErr) Add(p token.Pos) token.Pos {
	return d.pos + p
}

func ParseParamDecorator(cmt *ast.Comment, paramName string, paramType ast.Expr) (Decorator, *DecorationErr) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	x, _ := ParseExpr(text)
	xx, isCall := x.(*ast.CallExpr)
	if !isCall {
		return nil, nil
	}
	ident, hasName := xx.Fun.(*ast.Ident)
	if !hasName {
		return nil, nil
	}
	// path decor
	path, err := VerifyPathParamDecor(ident, xx.Args, paramName, paramType)
	if err != nil {
		return nil, err
	}
	if path != nil {
		return path, nil
	}
	return nil, nil
}

func ParseFnDecorator(cmt *ast.Comment) (Decorator, *DecorationErr) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	x, _ := ParseExpr(text)
	xx, isCall := x.(*ast.CallExpr)
	if !isCall {
		return nil, nil
	}
	ident, hasName := xx.Fun.(*ast.Ident)
	if !hasName {
		return nil, nil
	}
	// handler decor
	path, err := VerifyHandlerDecor(ident, xx.Args)
	if err != nil {
		return nil, err
	}
	if path != nil {
		return path, nil
	}
	// description decor
	descrDecor, err := VerifyDescrDecor(ident, xx.Args)
	if err != nil {
		return nil, err
	}
	if descrDecor != nil {
		return descrDecor, nil
	}
	return nil, nil
}

type DeclDecorators struct {
	declName   string
	decorators map[decoratorName]Decorator
	params     map[string]*FieldDecorators
}
type FieldType string

func (ft FieldType) Star() bool {
	return strings.HasPrefix(string(ft), "*")
}

func (ft FieldType) WithoutStar() string {
	r, _ := strings.CutPrefix(string(ft), "*")
	return r
}

type FieldDecorators struct {
	fieldName  string
	fieldType  FieldType
	decorators map[decoratorName]Decorator
}

func CodeLines(lines ...string) (c string) {
	c = "\n"
	for _, l := range lines {
		c += l + "\n"
	}
	c += ""
	return
}
func (p *parser) ParseFnDecorators(fnComments *ast.CommentGroup, fnName *ast.Ident, fnParams, fnResults *ast.FieldList) (fd *DeclDecorators) {
	fnDecorators := map[decoratorName]Decorator{}
	var hasHandlerDecorator = false
	if fnComments != nil && len(fnComments.List) > 0 {
		for _, cmt := range fnComments.List {
			decor, err := ParseFnDecorator(cmt)
			if err != nil {
				p.error(err.Add(cmt.Slash), err.msg)
				continue
			}
			if !hasHandlerDecorator {
				_, hasHandlerDecorator = decor.(*HandlerDecor)
			}
			if decor != nil {
				fnDecorators[decor.decoratorName()] = decor
			}
		}
		if !hasHandlerDecorator {
			// dont bother to parse comments for this fn, as the required 'handler' decor is not found
			return nil
		}

		fd = &DeclDecorators{
			declName:   fnName.Name,
			decorators: fnDecorators,
			params:     map[string]*FieldDecorators{},
		}
		for _, param := range fnParams.List {
			if param.Comment == nil || len(param.Comment.List) == 0 {
				/*
					TODO(ondbyte) implement more descriptive errors with code examples personalized to match the fn we are parsing, like
					example := CodeLines(
						fmt.Sprintf(`// handler("%v")`,param.Names[0]),
						fmt.Sprintf(`func %v(`, fnName.Name),
						fmt.Sprintf(`	// path("%v")`,param.Names[0]),
						fmt.Sprintf(`	%v string,`,param.Names[0],),
						`){`,
						``,
						`}`,
					) */
				p.error(param.Pos(), fmt.Sprintf("this function has a %v decorator, so this param needs one of these decorators %v", HANDLER, []decoratorName{PATH}))
			} else {
				paramDecorators := map[string]Decorator{}
				for _, cmt := range param.Comment.List {
					decor, err := ParseParamDecorator(cmt, param.Names[0].Name, param.Type)
					if err != nil {
						p.error(err.Add(cmt.Slash), err.msg)
						continue
					}
					if decor != nil {
						paramDecorators[param.Names[0].Name] = decor
					}
				}
				if len(paramDecorators) == 0 {
					p.error(param.Pos(), fmt.Sprintf("this function has a %v decorator, so this param needs one of these decorators %v", HANDLER, []decoratorName{PATH}))
					continue
				}
				paramType, err := StringifiedType(param.Type)
				if err != nil {
					p.error(param.Pos(), err.Error())
				}
				fd.params[param.Names[0].Name] = &FieldDecorators{
					fieldName:  param.Names[0].Name,
					fieldType:  FieldType(paramType),
					decorators: fnDecorators,
				}
			}
		}
	}
	return
}

// generates a handler func to handle the routes using decorator details
func GenFuncSrc(dd *DeclDecorators) (string, *DecorationErr) {
	h, ok := dd.decorators[HANDLER].(*HandlerDecor)
	if !ok {
		fmt.Println("no handler decor, no code will be generated")
		return "", nil
	}
	src := strings.Builder{}
	src.WriteString(fmt.Sprintf(`
	
	mux.HandleFunc("%v %v",func(w http.ResponseWriter, r *http.Request) {`, h.HttpMethod, h.Path))
	src.WriteRune('\n')

	handlerCall := &strings.Builder{}
	handlerCall.WriteString(dd.declName)
	handlerCall.WriteRune('(')
	for _, param := range dd.params {
		for _, decor := range param.decorators {
			switch fieldDecor := decor.(type) {
			case *PathParamDecor:
				hasPathParam := h.PathParams[fieldDecor.PathParamName]
				if !hasPathParam {
					return "", &DecorationErr{
						pos: fieldDecor.Pos,
						msg: "this path parameter is not defined in the handler decorator",
					}
				}
				src.WriteString(fmt.Sprintf(`%v := r.PathValue("%v")`, fieldDecor.PathParamName, fieldDecor.PathParamName))
				src.WriteRune('\n')
				src.WriteString(fmt.Sprintf("%v := new(%v)", param.fieldName, param.fieldType.WithoutStar()))
				src.WriteRune('\n')
				src.WriteString(fmt.Sprintf(`err = json.Unmarshal([]byte(%v),%v)`, fieldDecor.PathParamName, param.fieldName))
				src.WriteRune('\n')
				src.WriteString(fmt.Sprintf(`if err!=nil{panic(err)}`))
				src.WriteRune('\n')
			}
		}
		if !param.fieldType.Star() {
			handlerCall.WriteRune('*')
		}
		handlerCall.WriteString(param.fieldName)
		handlerCall.WriteRune(',')
	}
	handlerCall.WriteRune(')')
	src.WriteString(handlerCall.String())
	src.WriteRune('}')
	src.WriteRune(')')
	return src.String(), nil
}
