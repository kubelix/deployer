package service

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "gitlab.com/klinkert.io/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"gitlab.com/klinkert.io/kubelix/deployer/pkg/config"
)

func (r *ReconcileService) ensureService(svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
	coreService, err := r.newServiceForService(svc)
	if err != nil {
		return err
	}

	serviceName := types.NamespacedName{Name: coreService.Name, Namespace: coreService.Namespace}
	if err := r.ensureObject(reqLogger, coreService, serviceName); err != nil {
		return fmt.Errorf("failed to handle service: %v", err)
	}

	return nil
}

func (r *ReconcileService) newServiceForService(svc *appsv1alpha1.Service) (*corev1.Service, error) {
	labels := r.makeLabels(svc)

	coreService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:    svc.Spec.Ports.ToServicePorts(),
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	if len(config.Config.CoreService.Annotations) > 0 {
		coreService.SetAnnotations(config.Config.Ingress.Annotations)
	}

	if err := controllerutil.SetControllerReference(svc, coreService, r.scheme); err != nil {
		return nil, err
	}

	return coreService, nil
}
