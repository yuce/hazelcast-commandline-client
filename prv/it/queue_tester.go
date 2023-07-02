package it

import (
	"context"
	"testing"

	hz "github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/prv/check"
)

func WithQueue(tcx TestContext, fn func(m *hz.Queue)) {
	name := NewUniqueObjectName("queue")
	ctx := context.Background()
	m := check.MustValue(tcx.Client.GetQueue(ctx, name))
	fn(m)
}

func QueueTester(t *testing.T, fn func(tcx TestContext, m *hz.Queue)) {
	tcx := TestContext{T: t}
	tcx.Tester(func(tcx TestContext) {
		WithQueue(tcx, func(m *hz.Queue) {
			fn(tcx, m)
		})
	})
}
