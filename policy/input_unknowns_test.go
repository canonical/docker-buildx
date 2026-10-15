package policy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputUnknownRefsNestedMaterials(t *testing.T) {
	inp := Input{
		Image: &Image{
			Provenance: &ImageProvenance{
				Materials: []Input{
					{
						Image: &Image{
							Provenance: &ImageProvenance{
								Materials: []Input{{}},
							},
						},
					},
				},
			},
		},
	}
	inp.setUnknowns([]string{"input.image.provenance"})
	inp.Image.Provenance.Materials[0].setUnknowns([]string{"input.image.hasProvenance"})
	inp.Image.Provenance.Materials[0].Image.Provenance.Materials[0].setUnknowns([]string{"input.git.commit"})

	require.Equal(t, []string{
		"input.image.provenance",
		"input.image.provenance.materials[0].image.hasProvenance",
		"input.image.provenance.materials[0].image.provenance.materials[0].git.commit",
	}, inp.Unknowns())
}
