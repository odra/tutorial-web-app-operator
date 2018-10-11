package openshift

import (
	"k8s.io/client-go/kubernetes"
	v1template "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	v14 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	v13 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
)

const (
	OS_TEMPLATE_GROUP = "template.openshift.io"
	OS_TEMPLATE_VERSION = "v1"
	OS_API_PATH = "/apis"
	OS_API_MIMETYPE = "application/json"
	OS_TEMPLATE_RESOURCE = "processedtemplates"
)

type OSClient struct {
	kubeClient kubernetes.Interface
	TmplHandler TemplateHandler
	ocDCClient v14.AppsV1Interface
	ocRouteClient v13.RouteV1Interface

}

type Template struct {
	namespace string
	restClient *rest.RESTClient
}

type TemplateOpt struct {
	ApiKind string
	ApiVersion string
	ApiPath string
}


type Resource struct {}

type TemplateHandler interface {
	Process(tmpl *v1template.Template, params map[string]string) ([]runtime.RawExtension, error)
	FillParams(tmpl *v1template.Template, params map[string]string)
}
