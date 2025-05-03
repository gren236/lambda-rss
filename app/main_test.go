package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_handler(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler(tt.args.ctx)
			require.NoError(t, err)
		})
	}
}
