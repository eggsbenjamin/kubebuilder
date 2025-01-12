package addon

import (
	"github.com/eggsbenjamin/kubebuilder/pkg/model"
)

type Plugin struct {
}

func (p *Plugin) Pipe(u *model.Universe) error {
	functions := []PluginFunc{
		ExampleManifest,
		ExampleChannel,
		ReplaceController,
		ReplaceTypes,
	}

	for _, fn := range functions {
		if err := fn(u); err != nil {
			return err
		}

	}

	return nil
}
