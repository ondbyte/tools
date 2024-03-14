package parser

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"
	"go/token"
)

type decoratorName string

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

type DeclDecorators struct {
	declName   string
	decorators map[decoratorName]Decorator
	params     map[*ast.Field]*FieldDecorators
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
	if fnComments == nil || len(fnComments.List) == 0 {
		return
	}
	fnDecorators := map[decoratorName]Decorator{}
	handlerCommentIndex := -1
	var hd *HandlerDecor

	for i, v := range fnComments.List {
		// first fn decorator should be a handler
		dc, err := NewDecorComment(v)
		if err != nil {
			p.error(err.pos, err.msg)
			return
		}
		hd, err = dc.VerifyHandlerDecor()
		if err != nil {
			p.error(err.pos, err.msg)
			continue
		}
		if hd == nil {
			// not a handler
			continue //until we find one
		}
		handlerCommentIndex = i
		fnDecorators[hd.decoratorName()] = hd
	}
	if handlerCommentIndex == -1 {
		return
	}

	fd = &DeclDecorators{
		declName:   fnName.Name,
		decorators: fnDecorators,
		params:     map[*ast.Field]*FieldDecorators{},
	}
	fnCommentsLeft := fnComments.List[handlerCommentIndex:]
	if len(fnCommentsLeft) > 0 {
		for _, v := range fnCommentsLeft {
			dc, err := NewDecorComment(v)
			if err != nil {
				p.error(err.pos, err.msg)
				return
			}
			if dc == nil {
				continue
			}
			var dd Decorator
			dd, err = dc.VerifyDescrDecor()
			if err != nil {
				p.error(err.pos, err.msg)
				return
			}
			if dd != nil {
				fd.decorators[dd.decoratorName()] = dd
			} else {
				p.error(dc.DecorName.Pos(), "unknown decor")
			}
		}
		for _, param := range fnParams.List {
			if param.Comment == nil || len(param.Comment.List) == 0 {
				p.error(param.Pos(), fmt.Sprintf("this function has a %v decorator, so this param needs one of these decorators %v", HANDLER, []decoratorName{PATH}))
			} else {
				paramDecorators := map[decoratorName]Decorator{}
				for _, cmt := range param.Comment.List {
					dc, err := NewDecorComment(cmt)
					if err != nil {
						p.error(err.pos, err.msg)
						continue
					}
					descrDecor, err := dc.VerifyDescrDecor()
					if err != nil {
						p.error(err.pos, err.msg)
						continue
					}
					if descrDecor != nil {
						paramDecorators[descrDecor.decoratorName()] = descrDecor
						continue
					}
					pathParamDecor, err := dc.VerifyPathParamDecor()
					if err != nil {
						p.error(err.pos, err.msg)
						continue
					}
					if pathParamDecor != nil {
						if !hd.PathParams[pathParamDecor.PathParamName] {
							p.error(dc.DecorName.Pos(), "this path param is not in the path")
							continue
						}
						paramDecorators[pathParamDecor.decoratorName()] = pathParamDecor
						continue
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
				fd.params[param] = &FieldDecorators{
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
