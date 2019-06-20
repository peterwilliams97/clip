package clip_test

import (
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unipdf/v3/common"
)

func TestVertex(t *testing.T) {
	p0 := clip.Point{0, 0}
	p1 := clip.Point{1, 0}
	p2 := clip.Point{-1, 0}
	v0 := clip.NewVertex(p0, 0, nil, nil)
	v1 := clip.NewVertex(p1, 1, nil, nil)
	v2 := clip.NewVertex(p2, 2, nil, nil)
	v0.Join(v2, v1)
	v1.Join(v0, v2)
	v2.Join(v1, v0)
	common.Log.Debug("v0=%p %s", v0, v0)
	common.Log.Debug("v1=%p %s", v1, v1)
	common.Log.Debug("v2=%p %s", v2, v2)
	v0.Validate()
	v1.Validate()
	v2.Validate()
}
