package component_test

import (
	"context"
	"strings"
	"testing"

	"github.com/erigontech/erigon-lib/app"
	"github.com/erigontech/erigon-lib/app/component"
	liblog "github.com/erigontech/erigon-lib/log/v3"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

type provider struct {
}

func TestCreateComponent(t *testing.T) {
	c, err := component.NewComponent[provider](context.Background())
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:provider", c.Id().String())

	var p *provider = c.Provider()
	require.NotNil(t, p)

	c1, err := component.NewComponent[provider](context.Background(),
		component.WithId("my-id"))
	require.Nil(t, err)
	require.NotNil(t, c1)
	require.Equal(t, "root:my-id", c1.Id().String())

	c2, err := component.NewComponent[provider](context.Background(),
		component.WithId("my-id-2"),
		component.WithDependencies(c, c1))
	require.Nil(t, err)
	require.NotNil(t, c2)
	require.Equal(t, "root:my-id-2", c2.Id().String())
	require.Equal(t, "root:my-id", c1.Id().String())
	require.Equal(t, "root:provider", c.Id().String())
	require.True(t, c.HasDependent(c2))
	require.True(t, c1.HasDependent(c2))
	require.False(t, c.HasDependent(c1))
	require.False(t, c1.HasDependent(c))
}

func TestCreateDomain(t *testing.T) {
	d, err := component.NewComponentDomain(context.Background(), "domain")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "root:domain", d.Id().String())
	d1, err := component.NewComponentDomain(context.Background(), "domain-1",
		component.WithDependentDomain(d))
	require.Nil(t, err)
	require.NotNil(t, d1)
	require.Equal(t, "domain:domain-1", d1.Id().String())
	d2, err := component.NewComponentDomain(context.Background(), "domain-2",
		component.WithDependentDomain(d1))
	require.Nil(t, err)
	require.NotNil(t, d2)
	require.Equal(t, "domain-1:domain-2", d2.Id().String())
}

func TestCreateComponentInDomain(t *testing.T) {
	d, err := component.NewComponentDomain(context.Background(), "domain")
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "root:domain", d.Id().String())

	c, err := component.NewComponent[provider](context.Background(),
		component.WithDomain(d))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "domain:provider", c.Id().String())

	var p *provider = c.Provider()
	require.NotNil(t, p)

	c1, err := component.NewComponent[provider](context.Background())
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:provider", c1.Id().String())

	d1, err := component.NewComponentDomain(context.Background(), "domain-1",
		component.WithDependencies(c1))
	require.Nil(t, err)
	require.NotNil(t, d1)
	require.Equal(t, "root:domain-1", d1.Id().String())
	require.Equal(t, "domain-1:provider", c1.Id().String())

	d2, err := component.NewComponentDomain(context.Background(), "domain-2",
		component.WithDependencies(c, c1))
	require.Nil(t, err)
	require.NotNil(t, d2)
	require.Equal(t, "root:domain-2", d2.Id().String())
	require.Equal(t, "domain-2:provider", c.Id().String())
	require.Equal(t, "domain-2:provider", c1.Id().String())
}

func mockProvider(ctrl *gomock.Controller, callCount int) *component.MockComponentProvider {
	p := component.NewMockComponentProvider(ctrl)
	p.EXPECT().
		Configure(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(callCount)
	p.EXPECT().
		Initialize(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(callCount)
	p.EXPECT().
		Recover(gomock.Any()).
		Return(nil).
		Times(callCount)
	p.EXPECT().
		Activate(gomock.Any()).
		Return(nil).
		Times(callCount)
	p.EXPECT().
		Deactivate(gomock.Any()).
		Return(nil).
		Times(callCount)
	return p
}
func TestComponentLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	c, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:mockcomponentprovider", c.Id().String())

	err = c.Activate(context.Background())
	require.Nil(t, err)

	state, err := c.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)

	err = c.Deactivate(context.Background())
	require.Nil(t, err)

	state, err = c.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)

	d, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "root:d", d.Id().String())

	c1, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("c1"),
		component.WithProvider(mockProvider(ctrl, 1)),
		component.WithDependencies(d))
	require.Nil(t, err)
	require.NotNil(t, c1)
	require.Equal(t, "root:c1", c1.Id().String())

	err = c1.Activate(context.Background())
	require.Nil(t, err)

	state, err = c1.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)
	require.Equal(t, component.Active, d.State())
	require.Equal(t, component.Deactivated, c.State())

	err = c1.Deactivate(context.Background())
	require.Nil(t, err)

	state, err = c1.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)
	require.Equal(t, component.Deactivated, d.State())
	require.Equal(t, component.Deactivated, c.State())

	d1, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d1"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d1)
	require.Equal(t, "root:d1", d1.Id().String())

	d2, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d2"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d2)
	require.Equal(t, "root:d2", d2.Id().String())

	d3, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d3"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d3)
	require.Equal(t, "root:d3", d3.Id().String())

	c2, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("c2"),
		component.WithProvider(mockProvider(ctrl, 1)),
		component.WithDependencies(d1, d2, d3))

	require.Nil(t, err)
	require.NotNil(t, c2)
	require.Equal(t, "root:c2", c2.Id().String())

	err = c2.Activate(context.Background())
	require.Nil(t, err)

	state, err = c2.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)
	require.Equal(t, component.Active, d1.State())
	require.Equal(t, component.Active, d2.State())
	require.Equal(t, component.Active, d3.State())
	require.Equal(t, component.Deactivated, d.State())
	require.Equal(t, component.Deactivated, c1.State())
	require.Equal(t, component.Deactivated, c.State())

	err = c2.Deactivate(context.Background())
	require.Nil(t, err)

	state, err = c2.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)
	require.Equal(t, component.Deactivated, d1.State())
	require.Equal(t, component.Deactivated, d2.State())
	require.Equal(t, component.Deactivated, d3.State())
	require.Equal(t, component.Deactivated, d.State())
	require.Equal(t, component.Deactivated, c1.State())
	require.Equal(t, component.Deactivated, c.State())
}

func TestConfigre(t *testing.T) {
	ctrl := gomock.NewController(t)
	p := component.NewMockConfigurable(ctrl)
	p.EXPECT().
		Configure(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(2)

	c, err := component.NewComponent[component.MockConfigurable](context.Background(),
		component.WithProvider(p))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:mockconfigurable", c.Id().String())

	err = c.Activate(context.Background())
	require.Nil(t, err)

	state, err := c.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)
	require.Equal(t, component.Active, c.State())

	err = c.Configure(context.Background())
	require.Nil(t, err)
	require.Equal(t, component.Active, c.State())

	err = c.Deactivate(context.Background())
	require.Nil(t, err)

	state, err = c.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)
	require.Equal(t, component.Deactivated, c.State())

	p1 := component.NewMockConfigurable(ctrl)
	p1.EXPECT().
		Configure(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	c1, err := component.NewComponent[component.MockConfigurable](context.Background(),
		component.WithProvider(p1))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:mockconfigurable", c.Id().String())

	err = c1.Configure(context.Background())
	require.Nil(t, err)
	require.Equal(t, component.Configured, c1.State())

	err = c1.Activate(context.Background())
	require.Nil(t, err)

	state, err = c1.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)
	require.Equal(t, component.Active, c1.State())
}

func TestDomainLifecycle(t *testing.T) {
	dom, err := component.NewComponentDomain(context.Background(), "domain")
	require.Nil(t, err)
	require.NotNil(t, dom)
	require.Equal(t, "root:domain", dom.Id().String())

	ctrl := gomock.NewController(t)
	c, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithDependentDomain(dom),
		component.WithId("c"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "domain:c", c.Id().String())

	d, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d)
	require.Equal(t, "root:d", d.Id().String())

	c1, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithDependentDomain(dom),
		component.WithId("c1"),
		component.WithProvider(mockProvider(ctrl, 1)),
		component.WithDependencies(d))
	require.Nil(t, err)
	require.NotNil(t, c1)
	require.Equal(t, "domain:c1", c1.Id().String())
	require.Equal(t, "domain:d", d.Id().String())

	d1, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d1"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d1)
	require.Equal(t, "root:d1", d1.Id().String())

	d2, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d2"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d2)
	require.Equal(t, "root:d2", d2.Id().String())

	d3, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithId("d3"),
		component.WithProvider(mockProvider(ctrl, 1)))
	require.Nil(t, err)
	require.NotNil(t, d3)
	require.Equal(t, "root:d3", d3.Id().String())

	c2, err := component.NewComponent[component.MockComponentProvider](context.Background(),
		component.WithDependentDomain(dom),
		component.WithId("c2"),
		component.WithProvider(mockProvider(ctrl, 1)),
		component.WithDependencies(d1, d2, d3))
	require.Nil(t, err)
	require.NotNil(t, c2)
	require.Equal(t, "domain:c2", c2.Id().String())
	require.Equal(t, "domain:d1", d1.Id().String())
	require.Equal(t, "domain:d2", d2.Id().String())
	require.Equal(t, "domain:d3", d3.Id().String())

	err = dom.Activate(context.Background())
	require.Nil(t, err)

	state, err := dom.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)
	require.Equal(t, component.Active, dom.State())
	require.Equal(t, component.Active, d1.State())
	require.Equal(t, component.Active, d2.State())
	require.Equal(t, component.Active, d3.State())
	require.Equal(t, component.Active, d.State())
	require.Equal(t, component.Active, c1.State())
	require.Equal(t, component.Active, c2.State())
	require.Equal(t, component.Active, c.State())

	err = dom.Deactivate(context.Background())
	require.Nil(t, err)

	state, err = dom.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)
	require.Equal(t, component.Deactivated, dom.State())
	require.Equal(t, component.Deactivated, d1.State())
	require.Equal(t, component.Deactivated, d2.State())
	require.Equal(t, component.Deactivated, d3.State())
	require.Equal(t, component.Deactivated, d.State())
	require.Equal(t, component.Deactivated, c1.State())
	require.Equal(t, component.Deactivated, c2.State())
	require.Equal(t, component.Deactivated, c.State())
}

func TestLogger(t *testing.T) {
	prev := liblog.Root().GetHandler()
	defer liblog.Root().SetHandler(prev)

	liblog.Root().SetHandler(
		liblog.FuncHandler(func(r *liblog.Record) error {
			require.True(t, strings.HasPrefix(r.Msg, "[component:root:provider]"))
			require.False(t, strings.HasSuffix(r.Msg, " "))
			return nil
		}))

	component.LogLevel(liblog.LvlTrace)

	c, err := component.NewComponent[provider](context.Background())
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:provider", c.Id().String())

	c.Activate(context.Background())
	c.AwaitState(context.Background(), component.Active)
}

type ctxprovider struct {
	t *testing.T
	c component.Component[ctxprovider]
}

func (p ctxprovider) Configure(ctx context.Context, options ...app.Option) error {
	cmp := component.ComponentValue[ctxprovider](ctx)
	require.Equal(p.t, cmp, p.c)
	return nil
}

func (p ctxprovider) Initialize(ctx context.Context, options ...app.Option) error {
	prev := liblog.Root().GetHandler()
	defer liblog.Root().SetHandler(prev)

	liblog.Root().SetHandler(
		liblog.FuncHandler(func(r *liblog.Record) error {
			require.True(p.t, strings.HasPrefix(r.Msg, "[component:root:c]"))
			require.Equal(p.t, "[component:root:c] initializing (cmp)", r.Msg)
			require.Equal(p.t, liblog.LvlInfo, r.Lvl)
			return nil
		}))

	cmp := component.ComponentValue[ctxprovider](ctx)
	log := app.CtxLogger(ctx)
	log.SetLevel(liblog.LvlInfo)
	require.Equal(p.t, cmp, p.c)
	log.Info("initializing (cmp)")

	llog := app.NewLogger(liblog.LvlWarn, []string{"component_test"}, nil)

	liblog.Root().SetHandler(
		liblog.FuncHandler(func(r *liblog.Record) error {
			require.True(p.t, strings.HasPrefix(r.Msg, "[component_test:root:c]"))
			require.Equal(p.t, "[component_test:root:c] initializing (ctx)", r.Msg)
			require.Equal(p.t, liblog.LvlInfo, r.Lvl)
			return nil
		}))

	log = llog.CtxLogger(ctx)
	log.Info("initializing (ctx)")
	return nil
}

func (p ctxprovider) Recover(ctx context.Context) error {
	cmp := component.ComponentValue[ctxprovider](ctx)
	require.Equal(p.t, cmp, p.c)
	return nil
}

func (p ctxprovider) Activate(ctx context.Context) error {
	cmp := component.ComponentValue[ctxprovider](ctx)
	require.Equal(p.t, cmp, p.c)
	return nil
}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := component.NewComponent[ctxprovider](ctx,
		component.WithId("c"),
		component.WithProvider(&ctxprovider{t: t}))
	require.Nil(t, err)
	require.NotNil(t, c)
	require.Equal(t, "root:c", c.Id().String())

	c.Provider().c = c

	c.Configure(context.Background())

	c.Activate(context.Background())
	state, err := c.AwaitState(context.Background(), component.Active)
	require.Nil(t, err)
	require.Equal(t, component.Active, state)

	cancel()
	state, err = c.AwaitState(context.Background(), component.Deactivated)
	require.Nil(t, err)
	require.Equal(t, component.Deactivated, state)
	require.Equal(t, component.Deactivated, c.State())
}

func TestFlags(t *testing.T) {
}
func TestEvents(t *testing.T) {
}

func TestMultipleDependents(t *testing.T) {
}

func TestAddRemoveDeps(t *testing.T) {
}
