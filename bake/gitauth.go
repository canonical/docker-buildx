package bake

import (
	"os"
	"strings"

	"github.com/docker/buildx/util/buildflags"
	"github.com/docker/buildx/util/urlutil"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/gitutil"
)

const (
	bakeGitAuthTokenEnv  = "BUILDX_BAKE_GIT_AUTH_TOKEN" // #nosec G101 -- environment variable key, not a credential
	bakeGitAuthHeaderEnv = "BUILDX_BAKE_GIT_AUTH_HEADER"
)

func gitAuthSecretsFromEnv(remoteURL string) buildflags.Secrets {
	return gitAuthSecretsFromEnviron(os.Environ(), remoteURL)
}

func gitAuthSecretsFromEnviron(environ []string, remoteURL string) buildflags.Secrets {
	host, ok := gitAuthHostFromURL(remoteURL)
	if !ok {
		return nil
	}
	secrets := make(buildflags.Secrets, 0, 2)
	secrets = append(secrets, gitAuthSecretsForEnv(llb.GitAuthTokenKey, bakeGitAuthTokenEnv, environ, host)...)
	secrets = append(secrets, gitAuthSecretsForEnv(llb.GitAuthHeaderKey, bakeGitAuthHeaderEnv, environ, host)...)
	return secrets
}

func gitAuthSecretsForEnv(secretIDPrefix, envPrefix string, environ []string, host string) buildflags.Secrets {
	envKey, ok := findGitAuthEnvKey(envPrefix, environ)
	if !ok || host == "" {
		return nil
	}
	return buildflags.Secrets{&buildflags.Secret{
		ID:  secretIDPrefix + "." + host,
		Env: envKey,
	}}
}

func shouldAttachGitAuthSecrets(inputURL, contextPath string) bool {
	if !urlutil.IsRemoteURL(inputURL) || !urlutil.IsRemoteURL(contextPath) {
		return false
	}
	inputHost, ok := gitAuthHostFromURL(inputURL)
	if !ok {
		return false
	}
	contextHost, ok := gitAuthHostFromURL(contextPath)
	if !ok {
		return false
	}
	return strings.EqualFold(inputHost, contextHost)
}

func gitAuthHostFromURL(remoteURL string) (string, bool) {
	gitURL, err := gitutil.ParseURL(remoteURL)
	if err != nil || gitURL.Host == "" {
		return "", false
	}
	return gitURL.Host, true
}

func findGitAuthEnvKey(envKey string, environ []string) (string, bool) {
	for _, env := range environ {
		key, _, ok := strings.Cut(env, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(key, envKey) {
			return key, true
		}
	}
	return "", false
}
