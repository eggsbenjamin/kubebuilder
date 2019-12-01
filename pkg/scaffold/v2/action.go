/*
Copyright 2018 The Kubernetes Authors.

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

package v2

import (
	"path/filepath"
	"strings"

	"github.com/gobuffalo/flect"

	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/input"
	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/resource"
	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/util"
	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/v2/bpmn"
)

// Action scaffolds a Action for a Resource
type Action struct {
	input.Input

	// Resource is the Resource to make the Action for
	Resource *resource.Resource

	// ResourcePackage is the package of the Resource
	ResourcePackage string

	// Plural is the plural lowercase of kind
	Plural string

	// Is the Group + "." + Domain for the Resource
	GroupDomain string

	// Is the bpmn definition from which to scaffold logic
	BPMNDefinition *bpmn.Definition

	// Formatted BPMNDefinition ExclusiveGateway names to scaffold ConditionFuncs
	ConditionFuncNames []string

	// Formatted BPMNDefinition Task names to scaffold Actions
	ActionNames []string
}

// GetInput implements input.File
func (a *Action) GetInput() (input.Input, error) {

	a.ResourcePackage, a.GroupDomain = util.GetResourceInfo(a.Resource, a.Repo, a.Domain)

	if a.Plural == "" {
		a.Plural = flect.Pluralize(strings.ToLower(a.Resource.Kind))
	}

	if a.Path == "" {
		a.Path = filepath.Join("controllers",
			strings.ToLower(a.Resource.Kind)+"_actions.go")
	}

	for _, egw := range a.BPMNDefinition.Process.ExclusiveGateways {
		name := egw.Name
		if name == "" {
			name = egw.ID
		}

		a.ConditionFuncNames = append(a.ConditionFuncNames, strings.ReplaceAll(name, " ", ""))
	}

	for _, task := range a.BPMNDefinition.Process.Tasks {
		name := task.Name
		if name == "" {
			name = task.ID
		}

		a.ActionNames = append(a.ActionNames, strings.ReplaceAll(name, " ", ""))
	}

	a.TemplateBody = actionTemplate

	a.Input.IfExistsAction = input.Error
	return a.Input, nil
}

const actionTemplate = `{{ .Boilerplate }}

package controllers

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConditionFuncs 

// {{ .Resource.Kind }}ConditionFunc determines if a condition is met based on current cluster state
type {{ .Resource.Kind }}ConditionFunc func({{ .Resource.Kind }}ClusterState) bool

{{ range .ConditionFuncNames }}
	func {{ . }}(clusterState {{ $.Resource.Kind }}ClusterState) bool {
		// implement condition
		return false
	}
{{ end }}

// Actions

// {{ .Resource.Kind }}Action defines an interface for performing an arbitrary action on a {{ .Resource.Kind }}
type {{ .Resource.Kind }}Action interface {
	Name() string
	Execute({{ .Resource.Kind }}ClusterState) error
}

{{ range .ActionNames }}
type {{ . }}Action struct {
	k8sClient          client.Client
	// add other action dependencies
}

func (a *{{ . }}Action) Execute(state {{ $.Resource.Kind }}ClusterState) error {
	return nil
}

func (a *{{ . }}Action) Name() string {
	return "{{ . }}"
}

{{ end }}
	
type {{ .Resource.Kind }}ActionIdentifier struct {
	scheme             *runtime.Scheme
	k8sClient          client.Client
}

type {{ .Resource.Kind }}ActionIdentifierCfg struct {
	Scheme             *runtime.Scheme
	K8sClient          client.Client
}

func (c *{{ .Resource.Kind }}ActionIdentifier) Identify{{ .Resource.Kind }}Action(input Identify{{ .Resource.Kind }}ActionInput) ({{ .Resource.Kind }}Action, error) {
	return nil, nil
}
`
