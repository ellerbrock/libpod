// +build linux

package libpod

import (
	"context"

	"github.com/containers/libpod/libpod/image"
	"github.com/opencontainers/runtime-tools/generate"
)

const (
	// IDTruncLength is the length of the pod's id that will be used to make the
	// infra container name
	IDTruncLength = 12
)

func (r *Runtime) makeInfraContainer(ctx context.Context, p *Pod, imgName, imgID string) (*Container, error) {

	// Set up generator for infra container defaults
	g, err := generate.New("linux")
	if err != nil {
		return nil, err
	}

	g.SetRootReadonly(true)
	g.SetProcessArgs([]string{r.config.InfraCommand})

	containerName := p.ID()[:IDTruncLength] + "-infra"
	var options []CtrCreateOption
	options = append(options, r.WithPod(p))
	options = append(options, WithRootFSFromImage(imgID, imgName, false))
	options = append(options, WithName(containerName))
	options = append(options, withIsInfra())

	return r.newContainer(ctx, g.Config, options...)
}

// createInfraContainer wrap creates an infra container for a pod.
// An infra container becomes the basis for kernel namespace sharing between
// containers in the pod.
func (r *Runtime) createInfraContainer(ctx context.Context, p *Pod) (*Container, error) {
	if !r.valid {
		return nil, ErrRuntimeStopped
	}

	newImage, err := r.ImageRuntime().New(ctx, r.config.InfraImage, "", "", nil, nil, image.SigningOptions{}, false, false)
	if err != nil {
		return nil, err
	}

	data, err := newImage.Inspect(ctx)
	if err != nil {
		return nil, err
	}
	imageName := newImage.Names()[0]
	imageID := data.ID

	return r.makeInfraContainer(ctx, p, imageName, imageID)
}