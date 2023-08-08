package tests

import (
	"os"
	"testing"

	"github.com/docker/buildx/tests/workers"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/util/testutil/integration"
)

func init() {
	if integration.IsTestDockerd() {
		workers.InitDockerWorker()
		workers.InitDockerContainerWorker()
	} else {
		workers.InitRemoteWorker()
	}
}

func TestIntegration(t *testing.T) {
	var tests []func(t *testing.T, sb integration.Sandbox)
	tests = append(tests, buildTests...)
	tests = append(tests, bakeTests...)
	tests = append(tests, inspectTests...)
	tests = append(tests, lsTests...)
	testIntegration(t, tests...)
}

func testIntegration(t *testing.T, funcs ...func(t *testing.T, sb integration.Sandbox)) {
	mirroredImages := integration.OfficialImages("busybox:latest", "alpine:latest")
	buildkitImage := "docker.io/moby/buildkit:buildx-stable-1"
	if integration.IsTestDockerd() {
		if img, ok := os.LookupEnv("TEST_BUILDKIT_IMAGE"); ok {
			ref, err := reference.ParseNormalizedNamed(img)
			if err == nil {
				buildkitImage = ref.String()
			}
		}
	}
	mirroredImages["moby/buildkit:buildx-stable-1"] = buildkitImage
	mirrors := integration.WithMirroredImages(mirroredImages)

	tests := integration.TestFuncs(funcs...)
	integration.Run(t, tests, mirrors)
}
