package sourcemeta

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	gwclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	"github.com/stretchr/testify/require"
)

type fakeGatewayClient struct {
	metaCalls  atomic.Int32
	solveCalls atomic.Int32
	metaResp   *sourceresolver.MetaResponse
	metaErr    error
	solveRes   *gwclient.Result
	solveErr   error
}

func (f *fakeGatewayClient) ResolveSourceMetadata(context.Context, *pb.SourceOp, sourceresolver.Opt) (*sourceresolver.MetaResponse, error) {
	f.metaCalls.Add(1)
	return f.metaResp, f.metaErr
}

func (f *fakeGatewayClient) Solve(context.Context, gwclient.SolveRequest) (*gwclient.Result, error) {
	f.solveCalls.Add(1)
	if f.solveRes == nil && f.solveErr == nil {
		return nil, errors.New("unexpected solve call")
	}
	return f.solveRes, f.solveErr
}

func TestResolverCloseNoopBeforeResolve(t *testing.T) {
	t.Parallel()

	var called atomic.Int32
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		called.Add(1)
		return nil
	})

	require.NoError(t, r.Close())
	require.EqualValues(t, 0, called.Load())
}

func TestResolverResolveOpensOnce(t *testing.T) {
	t.Parallel()

	var runs atomic.Int32
	mr := &fakeGatewayClient{metaResp: &sourceresolver.MetaResponse{}}
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		runs.Add(1)
		ready <- mr
		<-ctx.Done()
		return context.Cause(ctx)
	})

	op := &pb.SourceOp{}
	_, err := r.ResolveSourceMetadata(t.Context(), op, sourceresolver.Opt{})
	require.NoError(t, err)
	_, err = r.ResolveSourceMetadata(t.Context(), op, sourceresolver.Opt{})
	require.NoError(t, err)

	require.EqualValues(t, 1, runs.Load())
	require.EqualValues(t, 2, mr.metaCalls.Load())
	require.NoError(t, r.Close())
}

func TestResolverCloseAfterOpenCancelsBuild(t *testing.T) {
	t.Parallel()

	var canceled atomic.Bool
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		ready <- &fakeGatewayClient{metaResp: &sourceresolver.MetaResponse{}}
		<-ctx.Done()
		canceled.Store(true)
		return context.Cause(ctx)
	})

	_, err := r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
	require.NoError(t, err)
	require.NoError(t, r.Close())
	require.True(t, canceled.Load())
}

func TestResolverOpenFailureIsSticky(t *testing.T) {
	t.Parallel()

	expected := errors.New("boom")
	var runs atomic.Int32
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		runs.Add(1)
		return expected
	})

	_, err := r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
	require.ErrorIs(t, err, expected)
	_, err = r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
	require.ErrorIs(t, err, expected)
	require.EqualValues(t, 1, runs.Load())
	require.ErrorIs(t, r.Close(), expected)
}

func TestResolverCloseIgnoresTerminalContextErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
	}{
		{name: "canceled", err: context.Canceled},
		{name: "deadline", err: context.DeadlineExceeded},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
				return tc.err
			})
			_, err := r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
			require.ErrorIs(t, err, tc.err)
			require.NoError(t, r.Close())
		})
	}
}

func TestResolverConcurrentResolveUsesSingleOpen(t *testing.T) {
	t.Parallel()

	var runs atomic.Int32
	mr := &fakeGatewayClient{metaResp: &sourceresolver.MetaResponse{}}
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		runs.Add(1)
		ready <- mr
		<-ctx.Done()
		return context.Cause(ctx)
	})

	const n = 16
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			_, err := r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, runs.Load())
	require.EqualValues(t, n, mr.metaCalls.Load())

	done := make(chan struct{})
	closeErr := make(chan error, 1)
	go func() {
		defer close(done)
		closeErr <- r.Close()
	}()
	select {
	case <-done:
		require.NoError(t, <-closeErr)
	case <-time.After(2 * time.Second):
		t.Fatal("close timed out")
	}
}

func TestResolverFirstCanceledContextDoesNotPoisonFutureCalls(t *testing.T) {
	t.Parallel()

	mr := &fakeGatewayClient{metaResp: &sourceresolver.MetaResponse{}}
	started := make(chan struct{})
	release := make(chan struct{})

	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		close(started)
		<-release
		ready <- mr
		<-ctx.Done()
		return context.Cause(ctx)
	})

	canceledCtx, cancel := context.WithCancelCause(t.Context())
	cancel(context.Canceled)
	_, err := r.ResolveSourceMetadata(canceledCtx, &pb.SourceOp{}, sourceresolver.Opt{})
	require.ErrorIs(t, err, context.Canceled)

	<-started
	close(release)

	_, err = r.ResolveSourceMetadata(t.Context(), &pb.SourceOp{}, sourceresolver.Opt{})
	require.NoError(t, err)
	require.NoError(t, r.Close())
}

func TestResolverResolveStateUsesSolve(t *testing.T) {
	t.Parallel()

	mr := &fakeGatewayClient{solveErr: errors.New("solve boom")}
	r := newWithRun(func(ctx context.Context, ready chan<- gatewayResolver) error {
		ready <- mr
		<-ctx.Done()
		return context.Cause(ctx)
	})

	st := llb.Scratch()
	_, err := r.ResolveState(t.Context(), st)
	require.EqualError(t, err, "solve boom")
	require.EqualValues(t, 1, mr.solveCalls.Load())
	require.NoError(t, r.Close())
}

var _ gatewayResolver = (*fakeGatewayClient)(nil)
var _ sourceresolver.MetaResolver = (*fakeGatewayClient)(nil)
