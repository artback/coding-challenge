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
			name: "small count",
			args: args{
				a: app,
				Parameters: Parameters{
					ip:     "192.168.0.1",
					count:  5,
					offset: 0,
				},
			},
		},
		{
			name: "small count",
			args: args{
				a: app,
				Parameters: Parameters{
					ip:     "192.168.0.1",
					count:  5,
					offset: 2,
				},
			},
		},
		{
			name: "large count",
			args: args{
				a: app,
				Parameters: Parameters{
					ip:     "192.168.0.1",
					count:  10000,
					offset: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, item := range GetContentItems(tt.args.a, tt.args.Parameters) {
				if Provider(item.Source) != DefaultConfig[(i+tt.args.offset)%len(DefaultConfig)].Type {
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
	param := Parameters{count: count, ip: ip, offset: offset}
	for i := 0; i < b.N; i++ {
		GetContentItems(app, param)
	}
}
func Benchmark_getContentItemsSmallCount(b *testing.B) {
	ip := "192.168.0.1"
	count := 5
	offset := 0
	param := Parameters{count: count, ip: ip, offset: offset}
	for i := 0; i < b.N; i++ {
		GetContentItems(app, param)
	}
}
