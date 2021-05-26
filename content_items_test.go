package main

import (
	"testing"
)

func Test_getContentItems(t *testing.T) {
	type args struct {
		a App
		Parameters
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "small Count",
			args: args{
				a: app,
				Parameters: Parameters{
					Ip:     "192.168.0.1",
					Count:  5,
					Offset: 0,
				},
			},
		},
		{
			name: "small Count",
			args: args{
				a: app,
				Parameters: Parameters{
					Ip:     "192.168.0.1",
					Count:  5,
					Offset: 2,
				},
			},
		},
		{
			name: "large Count",
			args: args{
				a: app,
				Parameters: Parameters{
					Ip:     "192.168.0.1",
					Count:  10000,
					Offset: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, item := range GetContentItems(tt.args.a, tt.args.Parameters) {
				if Provider(item.Source) != DefaultConfig[(i+tt.args.Offset)%len(DefaultConfig)].Type {
					t.Errorf(
						"Position %d: Got Provider %v instead of Provider %v",
						i, item.Source, DefaultConfig[i].Type,
					)
				}
			}
		})
	}
}

func Benchmark_getContentItemsLargeCount(b *testing.B) {
	ip := "192.168.0.1"
	count := 50000
	offset := 0
	param := Parameters{Count: count, Ip: ip, Offset: offset}
	for i := 0; i < b.N; i++ {
		GetContentItems(app, param)
	}
}
func Benchmark_getContentItemsSmallCount(b *testing.B) {
	ip := "192.168.0.1"
	count := 5
	offset := 0
	param := Parameters{Count: count, Ip: ip, Offset: offset}
	for i := 0; i < b.N; i++ {
		GetContentItems(app, param)
	}
}
