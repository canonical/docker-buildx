package policy

import (
	"errors"
	"testing"

	gwpb "github.com/moby/buildkit/frontend/gateway/pb"
	solverpb "github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/sourcepolicy/policysession"
	"github.com/stretchr/testify/require"
)

func TestPolicyIsPolicyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "matches-recorded-source",
			err:  errors.New("failed to solve: error evaluating the source policy: source \"docker-image://busybox:latest\" not allowed by policy: action DENY"),
			want: true,
		},
		{
			name: "does-not-match-without-buildkit-pattern",
			err:  errors.New("failed to parse dockerfile for docker-image://busybox:latest"),
			want: false,
		},
		{
			name: "does-not-match-unrelated-error",
			err:  errors.New("failed to solve: error evaluating the source policy: source \"docker-image://alpine:latest\" not allowed by policy: action DENY"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPolicy(Opt{})
			req := &policysession.CheckPolicyRequest{
				Source: &gwpb.ResolveSourceMetaResponse{
					Source: &solverpb.SourceOp{
						Identifier: "docker-image://busybox:latest",
					},
				},
			}
			p.recordDenyIdentifier(req)

			require.Equal(t, tt.want, p.IsPolicyError(tt.err))
		})
	}
}
