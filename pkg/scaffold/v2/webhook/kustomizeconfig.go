/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"path/filepath"

	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/input"
)

var _ input.File = &KustomizeConfigWebhook{}

// KustomizeConfigWebhook scaffolds the Kustomization file in manager folder.
type KustomizeConfigWebhook struct {
	input.Input
}

// GetInput implements input.File
func (c *KustomizeConfigWebhook) GetInput() (input.Input, error) {
	if c.Path == "" {
		c.Path = filepath.Join("config", "webhook", "kustomizeconfig.yaml")
	}
	c.TemplateBody = KustomizeConfigWebhookTemplate
	c.Input.IfExistsAction = input.Error
	return c.Input, nil
}

const KustomizeConfigWebhookTemplate = `# the following config is for teaching kustomize where to look at when substituting vars.
# It requires kustomize v2.1.0 or newer to work properly.
nameReference:
- kind: Service
  version: v1
  fieldSpecs:
  - kind: MutatingWebhookConfiguration
    group: admissionregistration.k8s.io
    path: webhooks/clientConfig/service/name
  - kind: ValidatingWebhookConfiguration
    group: admissionregistration.k8s.io
    path: webhooks/clientConfig/service/name

namespace:
- kind: MutatingWebhookConfiguration
  group: admissionregistration.k8s.io
  path: webhooks/clientConfig/service/namespace
  create: true
- kind: ValidatingWebhookConfiguration
  group: admissionregistration.k8s.io
  path: webhooks/clientConfig/service/namespace
  create: true

varReference:
- path: metadata/annotations
`
