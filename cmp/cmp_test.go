package cmp

import "testing"

func TestComparator_Compare(t *testing.T) {
	type fields struct {
		IsLeast   bool
		IsOrdered bool
	}
	type args struct {
		result interface{}
		expect interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base", fields{false, true}, args{
			map[string]interface{}{"name": "mike"},
			map[string]interface{}{"name": "mike"},
		}, false},
		{"base", fields{false, true}, args{
			map[string]interface{}{"name": "mike"},
			map[string]interface{}{"name": "igor"},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comparator{
				IsLeast:   tt.fields.IsLeast,
				IsOrdered: tt.fields.IsOrdered,
			}
			if err := c.Compare(tt.args.result, tt.args.expect); (err != nil) != tt.wantErr {
				t.Errorf("Comparator.Compare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComparator_cmpSlice(t *testing.T) {
	type fields struct {
		IsLeast   bool
		IsOrdered bool
	}
	type args struct {
		result []interface{}
		expect []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"equals", fields{false, true}, args{
			[]interface{}{5, "mike", 3.1},
			[]interface{}{5, "mike", 3.1},
		}, false},
		{"not equals", fields{false, true}, args{
			[]interface{}{5, "mike", 3.1},
			[]interface{}{5, "mike", 3.2},
		}, true},
		{"different lens", fields{false, true}, args{
			[]interface{}{5, "mike", 3.1, 2},
			[]interface{}{5, "mike", 3.1},
		}, true},

		{"is least", fields{true, true}, args{
			[]interface{}{5, "mike", 3.1, 2},
			[]interface{}{5, "mike", 3.1},
		}, false},

		{"not ordered", fields{false, false}, args{
			[]interface{}{"mike", 3.1, 5},
			[]interface{}{5, "mike", 3.1},
		}, false},

		{"not ordered, is least", fields{true, false}, args{
			[]interface{}{"mike", 3.1, 5, 2},
			[]interface{}{5, "mike", 3.1},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comparator{
				IsLeast:   tt.fields.IsLeast,
				IsOrdered: tt.fields.IsOrdered,
			}
			if err := c.cmpSlice(tt.args.result, tt.args.expect); (err != nil) != tt.wantErr {
				t.Errorf("Comparator.cmpSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComparator_CmpBody(t *testing.T) {
	type fields struct {
		IsRaw     bool
		IsLeast   bool
		IsOrdered bool
	}
	type args struct {
		r  []byte
		ex []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"", fields{IsRaw: true}, args{[]byte("hello"), []byte("hello")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comparator{
				IsRaw:     tt.fields.IsRaw,
				IsLeast:   tt.fields.IsLeast,
				IsOrdered: tt.fields.IsOrdered,
			}
			if err := c.CmpBody(tt.args.r, tt.args.ex); (err != nil) != tt.wantErr {
				t.Errorf("Comparator.CmpBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
