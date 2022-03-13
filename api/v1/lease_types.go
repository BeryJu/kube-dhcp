package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LeaseCommonSpec struct {
	OptionSet corev1.LocalObjectReference `json:"optionSet"`

	AddressLeaseTime string `json:"addressLeaseTime"`
}

// LeaseSpec defines the desired state of Lease
type LeaseSpec struct {
	LeaseCommonSpec `json:",inline"`
	Identifier      string                      `json:"identifier"`
	Address         string                      `json:"address"`
	Scope           corev1.LocalObjectReference `json:"scope"`
}

// LeaseStatus defines the observed state of Lease
type LeaseStatus struct {
	LastRequest string `json:"lastRequest"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.spec.address`
//+kubebuilder:printcolumn:name="Scope",type=string,JSONPath=`.spec.scope.name`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// Lease is the Schema for the leases API
type Lease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LeaseSpec   `json:"spec,omitempty"`
	Status LeaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LeaseList contains a list of Lease
type LeaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Lease{}, &LeaseList{})
}
