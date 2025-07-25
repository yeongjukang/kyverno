package image

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"
	"github.com/kyverno/kyverno/pkg/config"
)

type ImageInfo struct {
	// Registry is the URL address of the image registry e.g. `docker.io`
	Registry string `json:"registry,omitempty"`

	// Name is the image name portion e.g. `busybox`
	Name string `json:"name"`

	// Path is the repository path and image name e.g. `some-repository/busybox`
	Path string `json:"path"`

	// Tag is the image tag e.g. `v2`
	Tag string `json:"tag,omitempty"`

	// Digest is the image digest portion e.g. `sha256:128c6e3534b842a2eec139999b8ce8aa9a2af9907e2b9269550809d18cd832a3`
	Digest string `json:"digest,omitempty"`

	// Reference is an optional readable reference for the image e.g. `docker.io/nginx`
	Reference string `json:"reference,omitempty"`

	// ReferenceWithTag is an optional readable reference for the image along with the image tag e.g. `docker.io/nginx:v10`
	ReferenceWithTag string `json:"referenceWithTag,omitempty"`
}

func (i *ImageInfo) String() string {
	var image string
	if i.Registry != "" {
		image = fmt.Sprintf("%s/%s", i.Registry, i.Path)
	} else {
		image = i.Path
	}
	if i.Digest != "" {
		if i.Tag != "" {
			return fmt.Sprintf("%s:%s@%s", image, i.Tag, i.Digest)
		}
		return fmt.Sprintf("%s@%s", image, i.Digest)
	}
	return fmt.Sprintf("%s:%s", image, i.Tag)
}

func GetImageInfo(image string, cfg config.Configuration) (*ImageInfo, error) {
	// adding the default domain in order to properly parse image info
	fullImageName := addDefaultRegistry(image, cfg)
	ref, err := reference.Parse(fullImageName)
	if err != nil {
		return nil, fmt.Errorf("bad image: %s, defaultRegistry: %s, enableDefaultRegistryMutation: %t: %w", fullImageName, config.Configuration.GetDefaultRegistry(cfg), config.Configuration.GetEnableDefaultRegistryMutation(cfg), err)
	}

	var registry, path, name, tag, digest, referenceWithTag string
	if named, ok := ref.(reference.Named); ok {
		registry = reference.Domain(named)
		path = reference.Path(named)
		name = path[strings.LastIndex(path, "/")+1:]
	}

	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	}
	if digested, ok := ref.(reference.Digested); ok {
		digest = digested.Digest().String()
	}
	// set default tag - the domain is set via addDefaultRegistry before parsing
	if digest == "" && tag == "" {
		tag = "latest"
	}
	// if registry mutation isn't enabled don't add the default registry
	if fullImageName != image && !config.Configuration.GetEnableDefaultRegistryMutation(cfg) {
		registry = ""
	}

	if registry != "" {
		referenceWithTag = fmt.Sprintf("%s/%s:%s", registry, path, tag)
	} else {
		referenceWithTag = fmt.Sprintf("%s:%s", path, tag)
	}

	imageInfo := &ImageInfo{
		Registry:         registry,
		Name:             name,
		Path:             path,
		Tag:              tag,
		Digest:           digest,
		ReferenceWithTag: referenceWithTag,
	}

	imageInfo.Reference = imageInfo.String()
	return imageInfo, nil
}

// addDefaultRegistry always adds default registry
func addDefaultRegistry(name string, cfg config.Configuration) string {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost" && strings.ToLower(name[:i]) == name[:i]) {
		name = fmt.Sprintf("%s/%s", config.Configuration.GetDefaultRegistry(cfg), name)
	}
	return name
}
