package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go/ast"
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

var endpointRegex = regexp.MustCompile(`endpoint\("([^"]+)"\)`)
var endpointContentRegex = regexp.MustCompile(`{([^}]*)}`)

type Path struct {
	Raw    string
	Params map[string]ast.Expr
}

// decorator implements Decorator.
func (p *Path) decoratorName() decoratorName {
	return PATH
}

var pathRegex = regexp.MustCompile(`path\("([^"]+)"\)`)

type Query struct {
	Raw    string
	Params map[string]string
}

var queryRegex = regexp.MustCompile(`query\("([^"]+)"\)`)

type Header struct {
	Raw    string
	Params map[string]string
}

var headerRegex = regexp.MustCompile(`header\("([^"]+)"\)`)

type Body struct {
	Raw    string
	Params map[string]string
}

var bodyRegex = regexp.MustCompile(`body\("([^"]+)"\)`)

type DescriptionDecor struct {
	Data string
}

// decorator implements Decorator.
func (d *DescriptionDecor) decoratorName() decoratorName {
	return DESCR
}

type FuncDecorator struct {
	Raw    string
	Params map[string]bool
}

// decorator implements Decorator.
func (h *FuncDecorator) decoratorName() decoratorName {
	panic("unimplemented")
}

// handlerDecor implements HandlerDecorExpr.
func (d *DescriptionDecor) handlerDecor() {
	panic("unimplemented")
}

var descriptionRegex = regexp.MustCompile(`description\("([^"]+)"\)`)

// parses 'handle(...)'
func ParseHandler(ep *ast.CallExpr) (*FuncDecorator, error) {
	if i, ok := ep.Fun.(*ast.Ident); ok {
		if i.Name != "endpoint" {
			return nil, nil
		}
		if len(ep.Args) == 0 {
			return nil, fmt.Errorf("endpoint requires 1 string argument")
		}
		if i, ok := ep.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			matches := endpointContentRegex.FindAllStringSubmatch(us, -1)
			params := map[string]bool{}
			for _, match := range matches {
				if params[match[1]] {
					return nil, fmt.Errorf("two path params with same name %v", match[1])
				}
				params[match[1]] = true
			}
			return &FuncDecorator{
				Params: params,
			}, nil
		}
	}
	return nil, nil
}

func ProcessPath(pathX *ast.CallExpr, paramType ast.Expr) (*Path, error) {
	if i, ok := pathX.Fun.(*ast.Ident); ok {
		if i.Name != "path" {
			return nil, nil
		}
		if len(pathX.Args) == 0 {
			return nil, fmt.Errorf("path requires 1 string argument")
		}
		if i, ok := pathX.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			return &Path{
				Params: map[string]ast.Expr{
					us: paramType,
				},
			}, nil
		}
	}
	return nil, nil
}

func ParseDescr(ep *ast.CallExpr) (*DescriptionDecor, error) {
	if i, ok := ep.Fun.(*ast.Ident); ok {
		if i.Name != "description" {
			return nil, nil
		}
		if len(ep.Args) == 0 {
			return nil, fmt.Errorf("description requires 1 string argument")
		}
		if i, ok := ep.Args[0].(*ast.BasicLit); ok {
			us, err := strconv.Unquote(i.Value)
			if err != nil {
				return nil, fmt.Errorf("error unquoting string '%v' : %v", i.Value, err)
			}
			return &DescriptionDecor{
				Data: us,
			}, nil
		}
	}
	return nil, nil
}

func ParseParamDecorator(cmt *ast.Comment, paramType ast.Expr) (Decorator, error) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	if x, err := ParseExpr(text); err != nil {
		return nil, err
	} else if xx, ok := x.(*ast.CallExpr); !ok {
		return nil, nil
	} else if ep, err := ProcessPath(xx, paramType); err != nil {
		return nil, err
	} else if ep != nil {
		return ep, nil
	}
	return nil, nil
}

func ParseFnDecorator(cmt *ast.Comment) (Decorator, error) {
	text := cmt.Text
	text = strings.Trim(text, "/")
	if x, err := ParseExpr(text); err != nil {
		return nil, err
	} else if xx, ok := x.(*ast.CallExpr); !ok {
		return nil, nil
	} else if ep, err := ParseHandler(xx); err != nil {
		return nil, err
	} else if ep != nil {
		return ep, nil
	} else if descr, err := ParseDescr(xx); err != nil {
		return nil, err
	} else if descr != nil {
		return descr, nil
	}
	return nil, nil
}

type DeclDecorators struct {
	declName   string
	decorators map[decoratorName]Decorator
	params     map[string]*FieldDecorators
}

type FieldDecorators struct {
	fieldName  string
	fieldType  ast.Expr
	decorators map[decoratorName]Decorator
}

func (p *parser) ParseFnDecorators(fnName *ast.Ident, fnParams, fnResults *ast.FieldList) (fd *DeclDecorators) {
	fnDecorators := map[decoratorName]Decorator{}
	var isHandler = false
	if p.leadComment != nil && len(p.leadComment.List) > 0 {
		for _, cmt := range p.leadComment.List {
			decor, err := ParseFnDecorator(cmt)
			if err != nil {
				p.error(fnName.NamePos, err.Error())
				continue
			}
			if !isHandler {
				_, isHandler = decor.(*FuncDecorator)
				if !isHandler {
					// dont bother to parse comments for this fn, as the required 'handler' decor is not found
					return nil
				}
			}
			fnDecorators[decor.decoratorName()] = decor
		}
	}
	fd = &DeclDecorators{
		declName:   fd.declName,
		decorators: fnDecorators,
		params:     map[string]*FieldDecorators{},
	}
	for _, param := range fnParams.List {
		if param.Comment == nil || len(param.Comment.List) == 0 {
			p.error(param.Pos(), "needs a decorator/s")
		} else {
			var paramDecorators map[string]Decorator
			for _, cmt := range param.Comment.List {
				decor, err := ParseParamDecorator(cmt, param.Type)
				if err != nil {
					p.error(param.Pos(), err.Error())
					continue
				}
				paramDecorators[param.Names[0].Name] = decor
			}
			if len(paramDecorators) == 0 {
				p.error(param.Pos(), "found no decorator/s in your comments")
				continue
			}
			fd.params[param.Names[0].Name] = &FieldDecorators{
				fieldName:  param.Names[0].Name,
				fieldType:  param.Type,
				decorators: fnDecorators,
			}
		}
	}
	return
}
