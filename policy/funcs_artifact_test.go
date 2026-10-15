package policy

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	policyverifier "github.com/moby/policy-helpers"
	policytypes "github.com/moby/policy-helpers/types"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/opencontainers/go-digest"
	"github.com/sigstore/sigstore-go/pkg/fulcio/certificate"
	"github.com/stretchr/testify/require"
)

func TestBuiltinArtifactAttestationImpl(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		const bundlePath = "bundle.sigstore"
		bundleBytes := []byte("bundle-bytes")
		dgst := digest.FromString("artifact-bytes")
		st := &state{Input: Input{HTTP: &HTTP{Checksum: dgst.String()}}}

		expectedTS := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		expectedRaw := &policytypes.SignatureInfo{
			Kind:          policytypes.KindDockerGithubBuilder,
			SignatureType: policytypes.SignatureBundleV03,
			Signer: &certificate.Summary{
				CertificateIssuer:      "CN=sigstore-intermediate,O=sigstore.dev",
				SubjectAlternativeName: "https://github.com/docker/buildx/.github/workflows/release.yml@refs/tags/v0.31.1",
				Extensions: certificate.Extensions{
					Issuer:              "https://token.actions.githubusercontent.com",
					RunnerEnvironment:   "github-hosted",
					SourceRepositoryURI: "https://github.com/docker/buildx",
					SourceRepositoryRef: "refs/tags/v0.31.1",
				},
			},
			Timestamps: []policytypes.TimestampVerificationResult{{
				Type:      "rekor",
				URI:       "https://rekor.sigstore.dev",
				Timestamp: expectedTS,
			}},
		}

		p := NewPolicy(Opt{
			FS: func() (fs.StatFS, func() error, error) {
				return fstest.MapFS{
					bundlePath: &fstest.MapFile{Data: bundleBytes},
				}, func() error { return nil }, nil
			},
			VerifierProvider: func() (PolicyVerifier, error) {
				return &mockPolicyVerifier{
					verifyArtifact: func(_ context.Context, gotDigest digest.Digest, gotBundle []byte, _ ...policyverifier.ArtifactVerifyOpt) (*policytypes.SignatureInfo, error) {
						require.Equal(t, dgst, gotDigest)
						require.Equal(t, bundleBytes, gotBundle)
						return expectedRaw, nil
					},
				}, nil
			},
		})

		httpVal, err := ast.InterfaceToValue(st.Input.HTTP)
		require.NoError(t, err)

		term, err := p.builtinArtifactAttestationImpl(
			rego.BuiltinContext{Context: t.Context()},
			ast.NewTerm(httpVal),
			ast.StringTerm(bundlePath),
			st,
		)
		require.NoError(t, err)
		require.NotNil(t, term)

		expectedVal, err := ast.InterfaceToValue(toAttestationSignature(expectedRaw))
		require.NoError(t, err)
		require.Equal(t, 0, term.Value.Compare(expectedVal))
	})

	t.Run("verify failure returns undefined", func(t *testing.T) {
		const bundlePath = "bundle.sigstore"
		st := &state{Input: Input{HTTP: &HTTP{Checksum: digest.FromString("artifact-bytes").String()}}}

		p := NewPolicy(Opt{
			FS: func() (fs.StatFS, func() error, error) {
				return fstest.MapFS{bundlePath: &fstest.MapFile{Data: []byte("bundle")}}, func() error { return nil }, nil
			},
			VerifierProvider: func() (PolicyVerifier, error) {
				return &mockPolicyVerifier{
					verifyArtifact: func(context.Context, digest.Digest, []byte, ...policyverifier.ArtifactVerifyOpt) (*policytypes.SignatureInfo, error) {
						return nil, errors.New("verification failed")
					},
				}, nil
			},
		})

		httpVal, err := ast.InterfaceToValue(st.Input.HTTP)
		require.NoError(t, err)

		term, err := p.builtinArtifactAttestationImpl(
			rego.BuiltinContext{Context: t.Context()},
			ast.NewTerm(httpVal),
			ast.StringTerm(bundlePath),
			st,
		)
		require.NoError(t, err)
		require.Nil(t, term)
	})

	t.Run("missing checksum adds unknown", func(t *testing.T) {
		st := &state{Input: Input{HTTP: &HTTP{}}}

		p := NewPolicy(Opt{})
		httpVal, err := ast.InterfaceToValue(st.Input.HTTP)
		require.NoError(t, err)

		term, err := p.builtinArtifactAttestationImpl(
			rego.BuiltinContext{Context: t.Context()},
			ast.NewTerm(httpVal),
			ast.StringTerm("bundle.sigstore"),
			st,
		)
		require.NoError(t, err)
		require.Nil(t, term)
		require.Contains(t, st.Unknowns, funcArtifactAttestation)
	})
}

func TestBuiltinGithubAttestationImpl(t *testing.T) {
	t.Run("missing checksum adds unknown", func(t *testing.T) {
		st := &state{Input: Input{HTTP: &HTTP{}}}

		p := NewPolicy(Opt{})
		httpVal, err := ast.InterfaceToValue(st.Input.HTTP)
		require.NoError(t, err)

		term, err := p.builtinGithubAttestationImpl(
			rego.BuiltinContext{Context: t.Context()},
			ast.NewTerm(httpVal),
			ast.StringTerm("docker/buildx"),
			st,
		)
		require.NoError(t, err)
		require.Nil(t, term)
		require.Contains(t, st.Unknowns, funcGithubAttestation)
	})

	t.Run("resolver required", func(t *testing.T) {
		st := &state{Input: Input{HTTP: &HTTP{Checksum: digest.FromString("artifact-bytes").String()}}}

		p := NewPolicy(Opt{})
		httpVal, err := ast.InterfaceToValue(st.Input.HTTP)
		require.NoError(t, err)

		term, err := p.builtinGithubAttestationImpl(
			rego.BuiltinContext{Context: t.Context()},
			ast.NewTerm(httpVal),
			ast.StringTerm("docker/buildx"),
			st,
		)
		require.Nil(t, term)
		require.EqualError(t, err, "github_attestation: source resolver is not configured")
	})
}
