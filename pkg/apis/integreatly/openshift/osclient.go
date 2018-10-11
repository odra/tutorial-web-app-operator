package openshift

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	"errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/client-go/rest"
)

func NewOSClient(kubeClient kubernetes.Interface, kubeconfig *rest.Config) OSClient {
	routeClient, _ := routev1.NewForConfig(kubeconfig)
	dcClient, _ := appsv1.NewForConfig(kubeconfig)

	return OSClient{
		kubeClient:kubeClient,
		ocRouteClient:routeClient,
		ocDCClient:dcClient,
	}
}

func (osClient *OSClient) Bootstrap(namespace string) error {
	tmpl, err := NewTemplate(namespace)
	if err != nil {
		return  err
	}

	osClient.TmplHandler = tmpl

	return nil
}

func (osClient *OSClient) GetPod(ns string, dc string) (v1.Pod, error) {
	pods := osClient.kubeClient.CoreV1().Pods(ns)

	poList, err := pods.List(meta_v1.ListOptions{})
	if err != nil {
		return v1.Pod{}, err
	}

	for _, pod := range poList.Items {
		p := interface{}(pod).(v1.Pod)
		if val, ok := p.Labels["deploymentconfig"]; ok {
			if val == dc {
				return p, nil
			}
		}
	}

	return v1.Pod{}, errors.New("Pod not found")
}

func (osClient *OSClient) Delete(ns string, label string) error {
	deleteOpts := meta_v1.NewDeleteOptions(0)
	listOpts := meta_v1.ListOptions{LabelSelector: "app=" + label}

	err := osClient.ocDCClient.DeploymentConfigs(ns).DeleteCollection(deleteOpts, listOpts)
	if err != nil {
		return err
	}

	err = osClient.kubeClient.CoreV1().Services(ns).Delete(label, deleteOpts)
	if err != nil {
		return err
	}

	err = osClient.ocRouteClient.Routes(ns).DeleteCollection(deleteOpts, listOpts)
	if err != nil {
		return err
	}

	return nil
}
