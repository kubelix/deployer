package service

import (
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "github.com/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"github.com/kubelix/deployer/pkg/config"
	"github.com/kubelix/deployer/pkg/names"
)

func (r *ReconcileService) ensureDeployment(svc *appsv1alpha1.Service, dockerPullSecrets []*corev1.Secret, reqLogger logr.Logger) (*appsv1.Deployment, error) {
	dep, err := r.newDeploymentForService(svc, dockerPullSecrets)
	if err != nil {
		return nil, err
	}

	depName := types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}
	if err := r.ensureObject(reqLogger, svc, dep, depName); err != nil {
		return nil, fmt.Errorf("failed to handle deployment: %v", err)
	}

	return dep, nil
}

func (r *ReconcileService) newDeploymentForService(svc *appsv1alpha1.Service, dockerPullSecrets []*corev1.Secret) (*appsv1.Deployment, error) {
	labels := r.makeLabels(svc)
	filesConfigMapName := names.FormatDashFromParts(svc.Name, "files")

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: ptrInt32(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: svc.Spec.ServiceAccountName,
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										TopologyKey: "kubernetes.io/hostname",
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: r.makeKubelixLabels(svc),
										},
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: ptrInt64(30),
					ImagePullSecrets:              secretsToReferences(dockerPullSecrets),
					Containers: []corev1.Container{
						{
							ImagePullPolicy: corev1.PullAlways,
							Name:            svc.Name,
							Image:           svc.Spec.Image,
							Command:         svc.Spec.Command,
							Args:            svc.Spec.Args,
							Env:             svc.Spec.Env.ToEnvVars(),
							Resources:       svc.Spec.Resources,
							Ports:           svc.Spec.Ports.ToPodPorts(),
							VolumeMounts:    filesToVolumeMounts(svc),

							/**
							livenessProbe:
							  failureThreshold: 3
							  httpGet:
								path: /healthz
								port: app
								scheme: HTTP
							  periodSeconds: 10
							  successThreshold: 1
							  timeoutSeconds: 1
							readinessProbe:
							  failureThreshold: 3
							  httpGet:
								path: /healthz
								port: app
								scheme: HTTP
							  periodSeconds: 10
							  successThreshold: 1
							  timeoutSeconds: 1
							*/
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "files",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: filesConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if svc.Spec.Singleton {
		dep.Spec.Replicas = ptrOne
		dep.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
	} else {
		dep.Spec.Replicas = ptrThree
		dep.Spec.Strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
	}

	if len(config.Config.Deployment.Annotations) > 0 {
		dep.SetAnnotations(config.Config.Ingress.Annotations)
	}

	if err := controllerutil.SetControllerReference(svc, dep, r.scheme); err != nil {
		return nil, err
	}

	return dep, nil
}

func filesToVolumeMounts(svc *appsv1alpha1.Service) []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0)

	for _, file := range svc.Spec.Files {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "files",
			ReadOnly:  true,
			MountPath: file.Path,
			SubPath:   file.Name,
		})
	}

	return volumeMounts
}

func secretsToReferences(secrets []*corev1.Secret) []corev1.LocalObjectReference {
	refs := make([]corev1.LocalObjectReference, 0)
	for _, secret := range secrets {
		refs = append(refs, corev1.LocalObjectReference{Name: secret.Name})
	}
	return refs
}
