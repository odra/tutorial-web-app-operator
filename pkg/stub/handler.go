package stub

import (
	"context"

	"github.com/integr8ly/tutorial-web-app-operator/pkg/apis/integreatly/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/integr8ly/tutorial-web-app-operator/pkg/apis/integreatly/openshift"
	"github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/api/core/v1"
	"errors"
)

func NewHandler(m *Metrics, osClient openshift.OSClient, dynamicResourceClientFactory func(apiVersion, kind, namespace string) (dynamic.ResourceInterface, string, error)) sdk.Handler {
	return &Handler{
		metrics: m,
		osClient: osClient,
		dynamicResourceClientFactory: dynamicResourceClientFactory,
	}
}

type Metrics struct {
	operatorErrors prometheus.Counter
}

type Handler struct {
	// Metrics example
	metrics *Metrics
	osClient openshift.OSClient
	dynamicResourceClientFactory func(apiVersion, kind, namespace string) (dynamic.ResourceInterface, string, error)
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.WebApp:
		if o.GetDeletionTimestamp() != nil {
			err := h.delete(o)
			if err != nil {
				logrus.Errorf("Error deleting all operator related resources: %v", err)
				h.setStatus(err.Error(), o)
				return err
			}
			return nil
		}
		exts, err := h.processTemplate(o)
		if err != nil {
			logrus.Errorf("Error while processing the template: %v", err)
			h.setStatus(err.Error(), o)
			return err
		}

		runtimeObjs, err := h.getRuntimeObjs(exts)
		if err != nil {
			logrus.Errorf("Error parsing the runtime objects from the template: %v", err)
			h.setStatus(err.Error(), o)
			return err
		}

		err = h.provisonObjects(runtimeObjs, o)
		if err != nil {
			logrus.Errorf("Error provisioning the runtime objects: %v", err)
			h.setStatus(err.Error(), o)
			return err
		}

		if h.isAppReady(o) {
			h.setStatus("OK", o)
		} else {
			h.setStatus("", o)
		}

		return nil
	}
	return nil
}

func (h *Handler) delete(cr *v1alpha1.WebApp) error {
	ns := cr.Namespace
	appName :=  cr.Spec.DCName

	return h.osClient.Delete(ns, appName)
}

func (h *Handler) setStatus(msg string, cr *v1alpha1.WebApp) {
	cr.Status.Message = msg
	sdk.Update(cr)
}

func (h *Handler) processTemplate(cr *v1alpha1.WebApp) ([]runtime.RawExtension, error) {
	tmplPath := cr.Spec.TemplatePath
	res, err := openshift.LoadKubernetesResourceFromFile(tmplPath)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	params["OPENSHIFT_OAUTHCLIENT_ID"] = cr.Spec.OauthClientId
	params["OPENSHIFT_HOST"] = cr.Spec.OpenshiftHost

	tmpl := res.(*v1.Template)
	ext, err := h.osClient.TmplHandler.Process(tmpl, params)

	return ext, err
}

func (h *Handler) getRuntimeObjs(exts []runtime.RawExtension) ([]runtime.Object, error) {
	objects := make([]runtime.Object, 0)
	for _, ext := range exts {
		res, err := openshift.LoadKubernetesResource(ext.Raw)
		if err != nil {
			return nil, err
		}
		objects = append(objects, res)
	}

	return objects, nil
}

func (h *Handler) provisonObjects(objects []runtime.Object, cr *v1alpha1.WebApp) error {
	for _, o := range objects {
		gvk := o.GetObjectKind().GroupVersionKind()
		apiVersion, kind := gvk.ToAPIVersionAndKind()

		resourceClient, _, err := h.dynamicResourceClientFactory(apiVersion, kind, cr.Namespace)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to get resource client: %v", err))
		}

		unstructObj, err := k8sutil.UnstructuredFromRuntimeObject(o)
		if err != nil {
			return fmt.Errorf("%v failed to turn runtime object %s into unstructured object during provision", err, o.GetObjectKind().GroupVersionKind().String())
		}

		unstructObj, err = resourceClient.Create(unstructObj)
		if err != nil && !errors2.IsAlreadyExists(err) {
			return fmt.Errorf("%v failed to create object during provision with kind ", err, o.GetObjectKind().GroupVersionKind().String())
		}
	}

	return nil
}

func (h *Handler) isAppReady(cr *v1alpha1.WebApp) bool {
	pod, err := h.osClient.GetPod(cr.Namespace, cr.Spec.DCName)
	if err != nil {
		return  false
	}

	if pod.Status.Phase == corev1.PodRunning {
		return true
	}

	return false
}

func RegisterOperatorMetrics() (*Metrics, error) {
	operatorErrors := prometheus.NewCounter(prometheus.CounterOpts {
		Name: "integreatly_tutorial_webapp_operator_reconcile_errors_total",
		Help: "Number of errors that occurred while reconciling the integreatly tutorial webapp deployment",
	})

	err := prometheus.Register(operatorErrors)
	if err != nil {
		return nil, err
	}

	return &Metrics{operatorErrors: operatorErrors}, nil
}
