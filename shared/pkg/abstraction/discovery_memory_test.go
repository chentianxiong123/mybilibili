package abstraction

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryDiscovery_RegisterAndDiscover(t *testing.T) {
	d := newMemoryDiscovery()
	ctx := context.Background()

	err := d.Register(ctx, ServiceInstance{
		ServiceName: "core", InstanceID: "a", Host: "127.0.0.1", Port: 8080, Healthy: true,
	})
	require.NoError(t, err)
	err = d.Register(ctx, ServiceInstance{
		ServiceName: "core", InstanceID: "b", Host: "127.0.0.1", Port: 8081, Healthy: true,
	})
	require.NoError(t, err)

	insts, err := d.Discover(ctx, "core")
	require.NoError(t, err)
	assert.Len(t, insts, 2)
	assert.Equal(t, "a", insts[0].InstanceID)
	assert.Equal(t, 8081, insts[1].Port)
}

func TestMemoryDiscovery_DiscoverUnknownService(t *testing.T) {
	d := newMemoryDiscovery()
	insts, err := d.Discover(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Len(t, insts, 0)
}

func TestMemoryDiscovery_RegisterIsolationByService(t *testing.T) {
	d := newMemoryDiscovery()
	ctx := context.Background()
	require.NoError(t, d.Register(ctx, ServiceInstance{ServiceName: "core", InstanceID: "c1"}))
	require.NoError(t, d.Register(ctx, ServiceInstance{ServiceName: "search", InstanceID: "s1"}))

	core, _ := d.Discover(ctx, "core")
	search, _ := d.Discover(ctx, "search")
	assert.Len(t, core, 1)
	assert.Len(t, search, 1)
}

func TestMemoryDiscovery_DeregisterAndWatch(t *testing.T) {
	d := newMemoryDiscovery()
	ctx := context.Background()
	require.NoError(t, d.Register(ctx, ServiceInstance{ServiceName: "core", InstanceID: "a"}))
	require.NoError(t, d.Deregister(ctx))

	ch, err := d.Watch(ctx, "core")
	require.NoError(t, err)
	assert.NotNil(t, ch)

	assert.NoError(t, d.Healthy(ctx))
}
