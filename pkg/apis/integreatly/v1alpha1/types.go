package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []WebApp `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WebAppSpec   `json:"spec"`
	Status            WebAppStatus `json:"status,omitempty"`
}

type WebAppSpec struct {
	Image string `json:"image"`
	ImageTag string `json:"image_tag""`
	OauthClientId string `json:"oauth_client_id"`
	TemplatePath string `json:"template_path"`
	OpenshiftHost string `json:"openshift_host"`
	DCName string `json:"dc_name"`
}

type WebAppStatus struct {
	Message string `json:"message"`
}
