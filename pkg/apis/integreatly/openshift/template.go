package openshift

import (
	v1template "github.com/openshift/api/template/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"encoding/json"
	"fmt"
)

func NewTemplate(namespace string) (*Template, error) {
	inConfig := k8sclient.GetKubeConfig()
	config := rest.CopyConfig(inConfig)
	config.GroupVersion = &schema.GroupVersion {
		Group:   OS_TEMPLATE_GROUP,
		Version: OS_TEMPLATE_VERSION,
	}
	config.APIPath = OS_API_PATH
	config.AcceptContentTypes = OS_API_MIMETYPE
	config.ContentType = OS_API_MIMETYPE

	config.NegotiatedSerializer = basicNegotiatedSerializer{}
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &Template{namespace:namespace, restClient:restClient}, nil
}

func (template *Template) Process(tmpl *v1template.Template, params map[string]string) ([]runtime.RawExtension, error) {
	template.FillParams(tmpl, params)
	resource, err := json.Marshal(tmpl)
	if err != nil {
		return nil, err
	}

	result := template.restClient.
		Post().
		Namespace(template.namespace).
		Body(resource).
		Resource(OS_TEMPLATE_RESOURCE).
		Do()


	if result.Error() == nil {
		data, err := result.Raw()
		if err != nil {
			return nil, err
		}

		templ, err := LoadKubernetesResource(data)
		if err != nil {
			return nil, err
		}

		if v1Temp, ok := templ.(*v1template.Template); ok {
			return v1Temp.Objects, nil
		}

		return nil, fmt.Errorf("Wrong type returned by the server: %v",  templ)
	}


	return nil, result.Error()
}

func (template *Template) FillParams(tmpl *v1template.Template, params map[string]string) {
	for i, param := range tmpl.Parameters {
		if value, ok := params[param.Name]; ok {
			tmpl.Parameters[i].Value = value
		}
	}
}
