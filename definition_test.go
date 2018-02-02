package bdb

import (
	"reflect"
	"testing"
)

type ItemGroup struct {
	Oid  string
	Name string
}

func TestGetDefinition(t *testing.T) {
	type args struct {
		rType reflect.Type
	}
	tests := []struct {
		name       string
		args       args
		wantBTable BTable
		wantErr    bool
	}{
		{
			name: "test",
			args: args{reflect.TypeOf(ItemGroup{})},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBTable, err := GetDefinition(tt.args.rType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBTable, tt.wantBTable) {
				t.Errorf("GetDefinition() = %v, want %v", gotBTable, tt.wantBTable)
			}
		})
	}
}
