package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScopeSpec defines the desired state of Scope
type ScopeSpec struct {
	SubnetCIDR string `json:"subnetCIDR"`

	LeaseTemplate *LeaseCommonSpec `json:"leaseTemplate"`

	// +kubebuilder:default:="{{ .dhcp.HostName() }}"
	LeaseNameTemplate string `json:"leaseNameTemplate"`

	Default bool `json:"default"`
}

// ScopeStatus defines the observed state of Scope
type ScopeStatus struct {
	State string `json:"state"`

	UsedLeases int64 `json:"usedLeases"`
	FreeLeases int64 `json:"freeLeases"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Scope is the Schema for the scopes API
type Scope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScopeSpec   `json:"spec,omitempty"`
	Status ScopeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScopeList contains a list of Scope
type ScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scope `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Scope{}, &ScopeList{})
}
