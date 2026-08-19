package bake

import (
	"testing"

	"github.com/docker/buildx/util/buildflags"
	"github.com/moby/buildkit/client/llb"
	"github.com/stretchr/testify/require"
)

func TestGitAuthSecretsFromEnviron(t *testing.T) {
	t.Run("empty without remote url", func(t *testing.T) {
		secrets := gitAuthSecretsFromEnviron([]string{
			bakeGitAuthTokenEnv + "=token",
			bakeGitAuthHeaderEnv + "=basic",
		}, "")
		require.Empty(t, secrets)
	})
	t.Run("derives host from remote url", func(t *testing.T) {
		secrets := gitAuthSecretsFromEnviron([]string{
			bakeGitAuthTokenEnv + "=token",
			bakeGitAuthHeaderEnv + "=basic",
		}, "https://example.com/org/repo.git")
		require.Equal(t, []string{
			llb.GitAuthTokenKey + ".example.com|" + bakeGitAuthTokenEnv,
			llb.GitAuthHeaderKey + ".example.com|" + bakeGitAuthHeaderEnv,
		}, secretPairs(secrets))
	})
	t.Run("ignores host suffixed keys", func(t *testing.T) {
		secrets := gitAuthSecretsFromEnviron([]string{
			bakeGitAuthTokenEnv + ".example.com=token",
			bakeGitAuthHeaderEnv + ".example.com=basic",
		}, "https://example.com/org/repo.git")
		require.Empty(t, secrets)
	})
}

func TestShouldAttachGitAuthSecrets(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		contextPath string
		want        bool
	}{
		{
			name:        "same host http urls",
			inputURL:    "https://example.com/org/repo.git",
			contextPath: "https://example.com/another/repo.git",
			want:        true,
		},
		{
			name:        "same host mixed git url styles",
			inputURL:    "https://example.com/org/repo.git",
			contextPath: "git@example.com:another/repo.git",
			want:        true,
		},
		{
			name:        "different hosts",
			inputURL:    "https://example.com/org/repo.git",
			contextPath: "https://other.example.com/org/repo.git",
			want:        false,
		},
		{
			name:        "non remote context",
			inputURL:    "https://example.com/org/repo.git",
			contextPath: "cwd://src",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldAttachGitAuthSecrets(tt.inputURL, tt.contextPath))
		})
	}
}

func secretPairs(secrets buildflags.Secrets) []string {
	out := make([]string, 0, len(secrets))
	for _, s := range secrets {
		out = append(out, s.ID+"|"+s.Env)
	}
	return out
}
