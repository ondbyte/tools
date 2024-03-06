package parser

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestVerifyDecorArgs(t *testing.T) {
	type args struct {
		decorName         *ast.Ident
		requiredName      string
		args              []ast.Expr
		noOfRequiredParam int
	}
	tests := []struct {
		name  string
		args  args
		wantR []string
		wantD bool
	}{
		{
			name: "",
			args: args{
				decorName:    ast.NewIdent("handler"),
				requiredName: "handler",
				args: []ast.Expr{
					&ast.BasicLit{Value: "/users/{id}"},
				},
				noOfRequiredParam: 1,
			},
			wantR: []string{"/users/{id}"},
			wantD: false,
		},
		{
			name: "",
			args: args{
				decorName:    ast.NewIdent("hander"),
				requiredName: "handler",
				args: []ast.Expr{
					&ast.BasicLit{Value: "/users/{id}"},
				},
				noOfRequiredParam: 1,
			},
			wantR: nil,
			wantD: false,
		},
		{
			name: "",
			args: args{
				decorName:    ast.NewIdent("path"),
				requiredName: "path",
				args: []ast.Expr{
					&ast.BasicLit{Value: "1"},
					&ast.BasicLit{Value: "2"},
				},
				noOfRequiredParam: 1,
			},
			wantR: nil,
			wantD: true,
		},
		{
			name: "",
			args: args{
				decorName:    ast.NewIdent("path"),
				requiredName: "path",
				args: []ast.Expr{
					&ast.BasicLit{Value: "1"},
					&ast.BasicLit{Value: "2"},
				},
				noOfRequiredParam: 2,
			},
			wantR: []string{"1", "2"},
			wantD: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotD := VerifyDecorArgs(tt.args.decorName, tt.args.requiredName, tt.args.args, tt.args.noOfRequiredParam)
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("VerifyDecorArgs() gotR = %v, want %v", gotR, tt.wantR)
			}
			if tt.wantD && gotD == nil {
				t.Errorf("expected eror")
			}
		})
	}
}
