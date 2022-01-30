package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:validation:Enum=Initialized;Runing
type ClusterStatus string

const (
	ClusterInit    ClusterStatus = "Initialized"
	ClusterRunning ClusterStatus = "Runing"
)

// +genclient
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PatroniCluster struct {
	metav1.TypeMeta      `json:",inline"`
	metav1.ObjectMeta    `json:"metadata,omitempty"`
	PatroniClusterSpec   PatroniClusterStatus `json:"spec"`
	PatroniClusterStatus PatroniClusterStatus `json:"status,omitempty"`
}

type PatroniClusterSpec struct {
	NodeList []string `json:"nodeList"`
	Image    string   `json:"image"`
}

type PatroniClusterStatus struct {
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PatroniClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []*PatroniCluster `json:"items"`
}
