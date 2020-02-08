package service

import (
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "github.com/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"github.com/kubelix/deployer/pkg/config"
	"github.com/kubelix/deployer/pkg/names"
)

func (r *ReconcileService) ensureDockerPullSecrets(svc *appsv1alpha1.Service, reqLogger logr.Logger) ([]string, error) {
	secrets, err := r.newDockerPullSecretsForService(svc)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)

	for _, secret := range secrets {
		depName := types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}
		if err := r.ensureObject(reqLogger, svc, secret, depName); err != nil {
			return nil, fmt.Errorf("failed to handle secret: %v", err)
		}
		names = append(names, secret.Name)
	}
	return names, nil
}

func (r *ReconcileService) newDockerPullSecretsForService(svc *appsv1alpha1.Service) ([]*corev1.Secret, error) {
	labels := r.makeLabels(svc)
	secrets := make([]*corev1.Secret, 0)

	for _, reg := range config.Config.DockerPullSecretes {
		name := names.FormatDashFromParts(svc.Name, "docker-pull", reg.Registry)

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: svc.Namespace,
				Labels:    labels,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			StringData: map[string]string{
				corev1.DockerConfigJsonKey: formatDockerPullSecret(reg.Registry, reg.Username, reg.Password),
			},
		}

		if err := controllerutil.SetControllerReference(svc, secret, r.scheme); err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func formatDockerPullSecret(registry, username, password string) string {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", username, password)))
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(dockerConfigContent, registry, auth)))
}
