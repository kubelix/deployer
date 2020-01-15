package service

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "gitlab.com/klinkert.io/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"gitlab.com/klinkert.io/kubelix/deployer/pkg/names"
)

func (r *ReconcileService) ensureFilesConfigMap(svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
	config, err := r.newFilesConfigMapForService(svc)
	if err != nil {
		return err
	}

	depName := types.NamespacedName{Name: config.Name, Namespace: config.Namespace}
	if err := r.ensureObject(reqLogger, config, depName); err != nil {
		return fmt.Errorf("failed to handle secret: %v", err)
	}

	return nil
}

func (r *ReconcileService) newFilesConfigMapForService(svc *appsv1alpha1.Service) (*corev1.ConfigMap, error) {
	labels := r.makeLabels(svc)
	name := names.FormatDashFromParts(svc.Name, "files")

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{},
	}

	if err := controllerutil.SetControllerReference(svc, config, r.scheme); err != nil {
		return nil, err
	}

	for _, file := range svc.Spec.Files {
		if _, ok := config.Data[file.Name]; ok {
			return nil, fmt.Errorf("each file needs to have a unique name")
		}

		config.Data[file.Name] = file.Content
	}

	return config, nil
}
