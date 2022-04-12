package dingding

import (
	"testing"
)

func TestClient_SendRobotMessage(t *testing.T) {
	type args struct {
		user string
		msg  interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				user: "13681684970",
				msg: map[string]interface{}{
					"text":  "hello world",
					"title": "hello world",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New("dingxxxxxxxxx", "iLg4g32dxxxxxxxxx")
			userId, err := c.GetUserByMobile(tt.args.user)
			if err != nil {
				t.Errorf("Client.GetUserByMobile error = %v", err)
				return
			}
			if err := c.SendRobotMessage([]*string{&userId}, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Client.SendRobotMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
