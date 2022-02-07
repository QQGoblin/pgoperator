package cluster

import (
	"fmt"
	v1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"pgoperator/pkg/apis/cluster/v1alpha1"
)

func affinitySet(pClusterName string, require bool) coreV1.PodAntiAffinity {

	if require {
		return coreV1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []coreV1.PodAffinityTerm{
				{

					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"application":  "patroni",
							"cluster-name": pClusterName,
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	} else {

		return coreV1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []coreV1.WeightedPodAffinityTerm{
				{
					Weight: 1,
					PodAffinityTerm: coreV1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"application":  "patroni",
								"cluster-name": pClusterName,
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}

	}
}

func generatorStatefulset(indexName string, pCluster *v1alpha1.PatroniCluster) v1.StatefulSet {

	pClusterName := pCluster.Name
	statefulsetId := fmt.Sprintf("%s-%s", pCluster.Name, indexName)

	labelSelector := map[string]string{
		"application":    "patroni",
		"cluster-name":   pClusterName,
		"statefulset-id": statefulsetId,
	}

	podAffinitySet := affinitySet(pClusterName, pCluster.PatroniClusterSpec.RequirePodAntiAffinity)

	if pCluster.PatroniClusterSpec.ServiceAccount == "" {
		pCluster.PatroniClusterSpec.ServiceAccount = defaultServiceAccountName
	}

	var replicas int32 = 1
	var terminationGracePeriodSeconds int64 = 0

	return v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulsetId,
			Namespace: pCluster.Namespace,
			Labels:    labelSelector,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelSelector,
				},
				Spec: coreV1.PodSpec{
					Affinity: &coreV1.Affinity{
						PodAntiAffinity: &podAffinitySet,
					},
					Containers: []coreV1.Container{
						{
							Name:            "postgres",
							Image:           pCluster.PatroniClusterSpec.Image,
							ImagePullPolicy: coreV1.PullIfNotPresent,
							ReadinessProbe: &coreV1.Probe{
								ProbeHandler: coreV1.ProbeHandler{
									HTTPGet: &coreV1.HTTPGetAction{
										Path:   "/readiness",
										Port:   intstr.IntOrString{IntVal: 8008},
										Scheme: "HTTP",
									},
								},
								InitialDelaySeconds: 3,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Ports: []coreV1.ContainerPort{
								{
									ContainerPort: 8008,
									Protocol:      coreV1.ProtocolTCP,
								},
								{
									ContainerPort: 5432,
									Protocol:      coreV1.ProtocolTCP,
								},
							},
							Env: []coreV1.EnvVar{
								{
									Name: "PATRONI_KUBERNETES_POD_IP",
									ValueFrom: &coreV1.EnvVarSource{
										FieldRef: &coreV1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name: "PATRONI_KUBERNETES_NAMESPACE",
									ValueFrom: &coreV1.EnvVarSource{
										FieldRef: &coreV1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "PATRONI_KUBERNETES_BYPASS_API_SERVICE",
									Value: "true",
								},
								{
									Name:  "PATRONI_KUBERNETES_USE_ENDPOINTS",
									Value: "true",
								},
								{
									Name:  "PATRONI_KUBERNETES_LABELS",
									Value: fmt.Sprintf("{application: patroni, cluster-name: %s}", pClusterName),
								},
								{
									Name:  "PATRONI_SUPERUSER_USERNAME",
									Value: defaultSuperUserName,
								},
								{
									Name:  "PATRONI_SUPERUSER_PASSWORD",
									Value: defaultSuperUserPassword,
								},
								{
									Name:  "PATRONI_REPLICATION_USERNAME",
									Value: defaultReplicationUserName,
								},
								{
									Name:  "PATRONI_REPLICATION_PASSWORD",
									Value: defaultReplicationUserPassword,
								},
								{
									Name:  "PATRONI_SCOPE",
									Value: pClusterName,
								},
								{
									Name: "PATRONI_NAME",
									ValueFrom: &coreV1.EnvVarSource{
										FieldRef: &coreV1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "PATRONI_POSTGRESQL_DATA_DIR",
									Value: defaultPgDataPath,
								},
								{
									Name:  "PATRONI_POSTGRESQL_PGPASS",
									Value: defaultPgPass,
								},
								{
									Name:  "PATRONI_POSTGRESQL_LISTEN",
									Value: "0.0.0.0:5432",
								},
								{
									Name:  "PATRONI_RESTAPI_LISTEN",
									Value: "0.0.0.0:8008",
								},
								// TODO: sync 通过控制器统一设置
							},
							VolumeMounts: []coreV1.VolumeMount{
								{
									Name:      "pgdata",
									MountPath: "/home/postgres/pgdata",
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					ServiceAccountName:            pCluster.PatroniClusterSpec.ServiceAccount,
				},
			},
			VolumeClaimTemplates: []coreV1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"application":    "patroni",
							"statefulset-id": statefulsetId,
						},
						Name: "pgdata",
					},
					Spec: coreV1.PersistentVolumeClaimSpec{
						AccessModes: []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
						Resources: coreV1.ResourceRequirements{
							Requests: coreV1.ResourceList{
								coreV1.ResourceStorage: resource.MustParse("5Gi"),
							},
						},
					},
				},
			},
			ServiceName: fmt.Sprintf("%s-repl", pClusterName),
		},
	}
}
