package parser

import (
	"fmt"
	"go/format"
	"os"
	"testing"
)

func TestGenCode(t *testing.T) {
	type args struct {
		dd *DeclDecorators
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				dd: &DeclDecorators{
					declName: "HandleUsers",
					decorators: map[decoratorName]Decorator{
						HANDLER: &HandlerDecor{
							HttpMethod: "GET",
							Path:       "/users/{user_id}/orders/{order_id}",
							PathParams: map[string]bool{
								"user_id":  true,
								"order_id": true,
							},
						},
					},
					params: map[string]*FieldDecorators{
						"userId": {
							fieldName: "userId",
							fieldType: "int",
							decorators: map[decoratorName]Decorator{
								PATH: &PathParamDecor{
									PathParamName: "user_id",
								},
							},
						},
						"orderId": {
							fieldName: "orderId",
							fieldType: "*int",
							decorators: map[decoratorName]Decorator{
								PATH: &PathParamDecor{
									PathParamName: "order_id",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* if got := GenCode(tt.args.dd); got != tt.want {
				t.Errorf("GenCode() = %v, want %v", got, tt.want)
			} */
			got, dErr := GenFuncSrc(tt.args.dd)
			if dErr != nil {
				panic(dErr)
			}
			os.WriteFile("./gen.go", []byte(fmt.Sprintf("package parser\nfunc main(){%v}", got)), 0644)
			gotSrc, err := format.Source([]byte(got))
			fmt.Println(string(gotSrc), err)
		})
	}
}
