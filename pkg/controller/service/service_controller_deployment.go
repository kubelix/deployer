package service

import (
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "gitlab.com/klinkert.io/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"gitlab.com/klinkert.io/kubelix/deployer/pkg/config"
)

func (r *ReconcileService) ensureDeployment(svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
	dep, err := r.newDeploymentForService(svc)
	if err != nil {
		return err
	}

	depName := types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}
	if err := r.ensureObject(reqLogger, dep, depName); err != nil {
		return fmt.Errorf("failed to handle deployment: %v", err)
	}

	return nil
}

func (r *ReconcileService) newDeploymentForService(svc *appsv1alpha1.Service) (*appsv1.Deployment, error) {
	labels := r.makeLabels(svc)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			RevisionHistoryLimit: ptrInt32(10),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
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

							/**
								imagePullSecrets:
							      - name: nimca-auth-prod-docker-pull

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
