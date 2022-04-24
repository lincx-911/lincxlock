package main

import (
	"testing"
)

func BenchmarkCallZkLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CallLock("zklock", "/lincxlock/lock", []string{zkserver})
	}

	b.StopTimer()
}
func BenchmarkCallEtcdLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CallLock("etcdlock", "/lincxlock/lock", []string{etcdserver})
	}

	b.StopTimer()
}
func BenchmarkCallRedisLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CallLock("redislock", "lincxlock", []string{redisserver})
	}

	b.StopTimer()
}
