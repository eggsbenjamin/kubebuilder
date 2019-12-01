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
)

// Controller scaffolds a Controller for a Resource
type Controller struct {
	input.Input

	// Resource is the Resource to make the Controller for
	Resource *resource.Resource

	// ResourcePackage is the package of the Resource
	ResourcePackage string

	// Plural is the plural lowercase of kind
	Plural string

	// Is the Group + "." + Domain for the Resource
	GroupDomain string

	FromBPMN bool
}

// GetInput implements input.File
func (a *Controller) GetInput() (input.Input, error) {

	a.ResourcePackage, a.GroupDomain = util.GetResourceInfo(a.Resource, a.Repo, a.Domain)

	if a.Plural == "" {
		a.Plural = flect.Pluralize(strings.ToLower(a.Resource.Kind))
	}

	if a.Path == "" {
		a.Path = filepath.Join("controllers",
			strings.ToLower(a.Resource.Kind)+"_controller.go")
	}

	if a.FromBPMN {
		a.TemplateBody = controllerTemplateBPMN
	} else {
		a.TemplateBody = controllerTemplate
	}

	a.Input.IfExistsAction = input.Error
	return a.Input, nil
}

const controllerTemplate = `{{ .Boilerplate }}

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	{{ .Resource.GroupImportSafe }}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Version }}"
)

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Log r.Logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }}/status,verbs=get;update;patch

func (r *{{ .Resource.Kind }}Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("{{ .Resource.Kind | lower }}", req.NamespacedName)

	// your r.Logic here

	return ctrl.Result{}, nil
}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&{{ .Resource.GroupImportSafe }}{{ .Resource.Version }}.{{ .Resource.Kind }}{}).
		Complete(r)
}
`

const controllerTemplateBPMN = `{{ .Boilerplate }}

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	{{ .Resource.GroupImportSafe }}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Version }}"
)

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Log logr.Logger
	Scheme *runtime.Scheme
	actionIdentifier {{ .Resource.Kind }}ActionIdentifier
}

// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }}/status,verbs=get;update;patch
func (r *{{ .Resource.Kind }}Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("{{ .Resource.Kind | lower }}", req.NamespacedName)

	// your r.Logic here


	couchbaseRebalance := &{{ .Resource.GroupImportSafe }}{{ .Resource.Version }}.{{ .Resource.Kind }}{}
	err := r.Get(context.TODO(), req.NamespacedName, couchbaseRebalance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("{{ .Resource.Kind }} not found")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	clusterState, err := r.Gather{{ .Resource.Kind }}ClusterState(Gather{{ .Resource.Kind }}ClusterStateInput{
		{{ .Resource.Kind }}: couchbaseRebalance,
		Logger:             r.Log,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	action, err := r.actionIdentifier.Identify{{ .Resource.Kind }}Action(Identify{{ .Resource.Kind }}ActionInput{
		State:  clusterState,
		Logger: r.Log,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	if action != nil {
		err = action.Execute(clusterState)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

type {{ .Resource.Kind }}ClusterState struct {
	{{ .Resource.Kind }} *{{.Resource.GroupImportSafe }}{{ .Resource.Version }}.{{ .Resource.Kind }} 

	// declare relevant state here
}

type Gather{{ .Resource.Kind }}ClusterStateInput struct{
	{{ .Resource.Kind }} *{{.Resource.GroupImportSafe }}{{ .Resource.Version }}.{{ .Resource.Kind }} 
	Logger logr.Logger
}

func (r *{{ .Resource.Kind }}Reconciler) Gather{{ .Resource.Kind }}ClusterState(input Gather{{ .Resource.Kind }}ClusterStateInput) ({{ .Resource.Kind }}ClusterState, error) {

	// gather all relevant cluster state and return

	return {{ .Resource.Kind }}ClusterState{
		{{ .Resource.Kind }}: input.{{ .Resource.Kind }},
	}, nil
}

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&{{ .Resource.GroupImportSafe }}{{ .Resource.Version }}.{{ .Resource.Kind }}{}).
		Complete(r)
}
`
