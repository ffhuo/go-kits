package influx

import (
	"context"
	"testing"
	"time"
)

func TestClientAPI_Add(t *testing.T) {
	type fields struct {
		url       string
		authToken string
		org       string
		bucket    string
	}
	type args struct {
		table  string
		tags   map[string]string
		fields map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ClientAPI
	}{
		{
			name: "test_add",
			fields: fields{
				url:       "http://10.60.77.126:8086",
				authToken: "V3DKOVPUwlpTnhYvRFysTNSTajo5t-ZV4bNKrvqP1O7OPzkIwPmeZ6fxd1BNGyBEAofSG8_giNwY9iO-Nrg84w==",
				org:       "iot",
				bucket:    "billing",
			},
			args: args{
				table: "billing",
				tags: map[string]string{
					"serviceCode":  "AmazonEC2",
					"costType":     "EdpDiscount",
					"cloud":        "AWS",
					"payerAccount": "262089274283",
				},
				fields: map[string]interface{}{
					"cost": -0.0000000063000000771751274,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cli := New(tt.fields.url, tt.fields.authToken)
			api := cli.Set(tt.fields.org, tt.fields.bucket).WithContext(ctx)
			got := api.Measurement(tt.args.table).AddPoint(time.Now(), tt.args.tags, tt.args.fields)
			got.Flush()
		})
	}
}

func TestClientAPI_Query(t *testing.T) {
	type fields struct {
		url       string
		authToken string
		org       string
		bucket    string
	}
	type args struct {
		obj interface{}
		cmd string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test_add",
			fields: fields{
				url:       "http://10.60.77.126:8086",
				authToken: "V3DKOVPUwlpTnhYvRFysTNSTajo5t-ZV4bNKrvqP1O7OPzkIwPmeZ6fxd1BNGyBEAofSG8_giNwY9iO-Nrg84w==",
				org:       "iot",
				bucket:    "billing",
			},
			args: args{
				obj: []map[string]interface{}{},
				cmd: `from(bucket:"billing")
				|> range(start:2023-03-01T00:00:00Z,stop:2023-03-31T00:00:00Z)
				|> filter(fn: (r)=>r._field == "cost")
				|> aggregateWindow(every:1h, fn:sum)
				|> holtWinters(n:30,seasonality:0,interval:1h)`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cli := New(tt.fields.url, tt.fields.authToken)
			api := cli.Set(tt.fields.org, tt.fields.bucket).WithContext(ctx)
			if err := api.Query(tt.args.obj, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("ClientAPI.Query() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
