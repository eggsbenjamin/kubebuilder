package addon

import (
	"path/filepath"
	"strings"

	"github.com/eggsbenjamin/kubebuilder/pkg/model"
	"github.com/eggsbenjamin/kubebuilder/pkg/scaffold/input"
)

func ReplaceController(u *model.Universe) error {
	templateBody := controllerTemplate

	funcs := DefaultTemplateFunctions()
	contents, err := RunTemplate("controller", templateBody, u, funcs)
	if err != nil {
		return err
	}

	m := &model.File{
		Path:           filepath.Join("controllers", strings.ToLower(u.Resource.Kind)+"_controller.go"),
		Contents:       contents,
		IfExistsAction: input.Error,
	}

	ReplaceFileIfExists(u, m)

	return nil
}

const controllerTemplate = `{{ .Boilerplate }}

package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"github.com/eggsbenjamin/kubebuilder-declarative-pattern/pkg/patterns/addon"
	"github.com/eggsbenjamin/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/status"
	"github.com/eggsbenjamin/kubebuilder-declarative-pattern/pkg/patterns/declarative"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "{{ .Resource.GoPackage }}/{{ .Resource.Version }}"
)

var _ reconcile.Reconciler = &{{ .Resource.Kind }}Reconciler{}

// {{ .Resource.Kind }}Reconciler reconciles a {{ .Resource.Kind }} object
type {{ .Resource.Kind }}Reconciler struct {
	client.Client
	Log logr.Logger
	Scheme *runtime.Scheme

	declarative.Reconciler
}

// +kubebuilder:rbac:groups={{.Resource.GroupDomain}},resources={{ .Resource.Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{.Resource.GroupDomain}},resources={{ .Resource.Plural }}/status,verbs=get;update;patch

func (r *{{ .Resource.Kind }}Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	addon.Init()

	labels := map[string]string{
		"k8s-app": "{{ .Resource.Kind | lower }}",
	}

	watchLabels := declarative.SourceLabel(mgr.GetScheme())

	if err := r.Reconciler.Init(mgr, &api.{{ .Resource.Kind }}{},
		declarative.WithObjectTransform(declarative.AddLabels(labels)),
		declarative.WithOwner(declarative.SourceAsOwner),
		declarative.WithLabels(watchLabels),
		declarative.WithStatus(status.NewBasic(mgr.GetClient())),
		// TODO: add an application to your manifest:  declarative.WithObjectTransform(addon.TransformApplicationFromStatus),
		// TODO: add an application to your manifest:  declarative.WithManagedApplication(watchLabels),
		declarative.WithObjectTransform(addon.ApplyPatches),
	); err != nil {
		return err
	}

	c, err := controller.New("{{ .Resource.Kind | lower }}-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to {{ .Resource.Kind }}
	err = c.Watch(&source.Kind{Type: &api.{{ .Resource.Kind }}{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to deployed objects
	_, err = declarative.WatchAll(mgr.GetConfig(), c, r, watchLabels)
	if err != nil {
		return err
	}

	return nil
}
`
