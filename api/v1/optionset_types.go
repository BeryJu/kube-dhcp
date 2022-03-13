package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OptionSetSpec defines the desired state of OptionSet
type OptionSetSpec struct {
	Options []*Option `json:"options"`
}

type Option struct {
	// Raw tag sent to the client, see https://www.iana.org/assignments/bootp-dhcp-parameters/bootp-dhcp-parameters.xhtml
	// +kubebuilder:validation:Optional
	// +nullable
	Tag *uint8 `json:"tag"`
	// Tag name
	// +kubebuilder:validation:Optional
	// +nullable
	TagName *string  `json:"tagName"`
	Values  []string `json:"values"`

	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty" protobuf:"bytes,3,opt,name=configMapKeyRef"`
	// Selects a key of a secret in the pod's namespace
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

// OptionSetStatus defines the observed state of OptionSet
type OptionSetStatus struct {
	Valid bool `json:"valid"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OptionSet is the Schema for the optionsets API
type OptionSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OptionSetSpec   `json:"spec,omitempty"`
	Status OptionSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OptionSetList contains a list of OptionSet
type OptionSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OptionSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OptionSet{}, &OptionSetList{})
}
