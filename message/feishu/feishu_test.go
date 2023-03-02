package feishu

import "testing"

func TestClient_SendWebhookMessage(t *testing.T) {
	type args struct {
		appId     string
		appSecret string
		url       string
		subject   string
		content   string
		atAll     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				appId:     "",
				appSecret: "",
				url:       "",
				subject:   "",
				content:   "",
				atAll:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.args.appId, tt.args.appSecret)
			if err := client.SendWebhookMessage(tt.args.url, tt.args.subject, tt.args.content, tt.args.atAll); (err != nil) != tt.wantErr {
				t.Errorf("Client.SendWebhookMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
