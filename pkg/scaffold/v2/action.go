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
	"bytes"
	"fmt"
	"io"
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

	// RenderedActionIdentifierFunction pre rendered due to complexity
	RenderedActionIdentifierFunction string
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
		a.ConditionFuncNames = append(a.ConditionFuncNames, getConditionFuncName(egw))
	}

	for _, task := range a.BPMNDefinition.Process.Tasks {
		a.ActionNames = append(a.ActionNames, getActionName(task))
	}

	identifyActionFunctionBody, err := a.DFS(a.BPMNDefinition)
	if err != nil {
		return a.Input, fmt.Errorf("error rendering bpmn logical flow: %q", err)
	}

	a.RenderedActionIdentifierFunction = identifyActionFunctionBody

	a.TemplateBody = actionTemplate

	a.Input.IfExistsAction = input.Error
	return a.Input, nil
}

// DFS performs a depth first search on the bpmn process to render the logic.
func (a *Action) DFS(def *bpmn.Definition) (string, error) {
	if def.Process.StartEvent.ID == "" {
		return "", fmt.Errorf("missing StartEvent")
	}

	out := &bytes.Buffer{}

	fmt.Fprintf(out, "func (i *%[1]sActionIdentifier) Identify%[1]sAction(input Identify%[1]sActionInput) (%[1]sAction, error) {\n", a.Resource.Kind)

	visited := map[string]struct{}{}
	a.DFSInner(def, def.Process.StartEvent.ID, visited, out)

	fmt.Fprintf(out, "\n}\n")

	return out.String(), nil
}

func (a *Action) DFSInner(def *bpmn.Definition, v string, visited map[string]struct{}, out io.Writer) error {
	visited[v] = struct{}{}
	elements := []bpmn.Element{}
	terminalElements := []bpmn.Element{}

	for _, v2 := range def.Process.DAG.AdjacenyList[v] {
		if _, ok := visited[v2]; ok {
			return nil // no more vertices to traverse
		}

		elem, ok := def.Process.GetElement(v2)
		if !ok {
			fmt.Errorf("unable to get element %v", v2)
		}

		switch elem.Type() {
		case bpmn.ElementTypeExclusiveGateway:
			elements = append(elements, elem)
		case bpmn.ElementTypeTask:
			terminalElements = append(terminalElements, elem)
		case bpmn.ElementTypeEnd:
			terminalElements = append(terminalElements, elem)
		default:
			return fmt.Errorf("unexpected element type %s", elem.Type())
		}
	}

	if len(terminalElements) == 0 {
		terminalElements = append(terminalElements, bpmn.EndEvent{}) // implicit end
	}

	if len(terminalElements) > 1 {
		return fmt.Errorf("unable to execute actions concurrently. Element %s can only have a single TaskEvent or EndEvent element as a direct child")
	}

	elements = append(elements, terminalElements...) // we know that there is a a single terminal element and it must come after the others

	for _, elem := range elements {
		switch elem.Type() {
		case bpmn.ElementTypeExclusiveGateway:
			egw := elem.(bpmn.ExclusiveGateway)
			fmt.Fprintf(out, "\nif %s(input.State) {\n", getConditionFuncName(egw))
			a.DFSInner(def, egw.ID, visited, out)
		case bpmn.ElementTypeTask:
			fmt.Fprintf(out, "\nreturn &%sAction{}, nil", getActionName(elem.(bpmn.Task)))
		case bpmn.ElementTypeEnd:
			fmt.Fprintf(out, "\nreturn nil, nil")
		}
	}

	if v != def.Process.StartEvent.ID {
		fmt.Fprintf(out, "\n}\n")
	}

	return nil
}

func getConditionFuncName(c bpmn.ExclusiveGateway) string {
	name := c.Name
	if name == "" {
		name = c.ID
	}

	return strings.ReplaceAll(name, " ", "")
}

func getActionName(t bpmn.Task) string {
	name := t.Name
	if name == "" {
		name = t.ID
	}

	return strings.ReplaceAll(name, " ", "")
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
		// implement condition logic
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
	// implement action logic
	return nil
}

func (a *{{ . }}Action) Name() string {
	return "{{ . }}"
}

{{ end }}

// Action Identifier
	
type {{ .Resource.Kind }}ActionIdentifier struct {
	scheme             *runtime.Scheme
	k8sClient          client.Client
}

type {{ .Resource.Kind }}ActionIdentifierCfg struct {
	Scheme             *runtime.Scheme
	K8sClient          client.Client
}

type Identify{{ .Resource.Kind }}ActionInput struct {
	State {{ .Resource.Kind }}ClusterState
	Logger logr.Logger
}

// AUTOGENERATED FROM BPMN.

{{ .RenderedActionIdentifierFunction }}
`
