package parser

import (
	"testing"
)

func TestStringifiedType(t *testing.T) {
	type args struct {
		x string
	}
	tests := []struct {
		name    string
		args    args
		wantT   string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				x: "int",
			},
			wantT:   "int",
			wantErr: false,
		},
		{
			name: "",
			args: args{
				x: "*int",
			},
			wantT:   "*int",
			wantErr: false,
		},
		{
			name: "",
			args: args{
				x: "**int",
			},
			wantT:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, _ := ParseExpr(tt.args.x)
			gotT, err := StringifiedType(x)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringifiedType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotT != tt.wantT {
				t.Errorf("StringifiedType() = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}
