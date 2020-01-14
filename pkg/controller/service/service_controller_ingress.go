package service

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha1 "gitlab.com/klinkert.io/kubelix/deployer/pkg/apis/apps/v1alpha1"
	"gitlab.com/klinkert.io/kubelix/deployer/pkg/config"
	"gitlab.com/klinkert.io/kubelix/deployer/pkg/names"
)

func (r *ReconcileService) ensureIngresses(err error, svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
	ingresses, err := r.newIngressesForService(svc)
	if err != nil {
		return err
	}

	for _, ingress := range ingresses {
		depName := types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}
		if err := r.ensureObject(reqLogger, ingress, depName); err != nil {
			return fmt.Errorf("failed to handle ingress: %v", err)
		}
	}

	return nil
}

func (r *ReconcileService) newIngressesForService(svc *appsv1alpha1.Service) ([]*networkingv1beta1.Ingress, error) {
	labels := r.makeLabels(svc)
	ingresses := make([]*networkingv1beta1.Ingress, 0)

	for _, p := range svc.Spec.Ports {
		if len(p.Ingresses) == 0 {
			continue
		}

		for _, ing := range p.Ingresses {
			name := names.FormatDash(strings.Join([]string{svc.Name, p.Name, ing.Host}, "-"))

			ingress := &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: svc.Namespace,
					Labels:    labels,
				},
				Spec: networkingv1beta1.IngressSpec{
					Rules: r.makeIngressRules(svc, p, ing),
					TLS: []networkingv1beta1.IngressTLS{
						{
							Hosts:      []string{ing.Host},
							SecretName: name + "-tls",
						},
					},
				},
			}

			if len(config.Config.Ingress.Annotations) > 0 {
				ingress.SetAnnotations(config.Config.Ingress.Annotations)
			}

			if err := controllerutil.SetControllerReference(svc, ingress, r.scheme); err != nil {
				return nil, err
			}

			ingresses = append(ingresses, ingress)
		}
	}

	return ingresses, nil
}

func (r *ReconcileService) makeIngressRules(svc *appsv1alpha1.Service, p appsv1alpha1.Port, ing appsv1alpha1.PortIngress) []networkingv1beta1.IngressRule {
	rules := make([]networkingv1beta1.IngressRule, 0)
	paths := make([]networkingv1beta1.HTTPIngressPath, 0)

	if len(ing.Paths) == 0 {
		ing.Paths = []string{"/"}
	}

	for _, path := range ing.Paths {
		paths = append(paths, networkingv1beta1.HTTPIngressPath{
			Path: path,
			Backend: networkingv1beta1.IngressBackend{
				ServicePort: intstr.FromString(p.Name),
				ServiceName: svc.Name,
			},
		})
	}

	rules = append(rules, networkingv1beta1.IngressRule{
		Host: ing.Host,
		IngressRuleValue: networkingv1beta1.IngressRuleValue{
			HTTP: &networkingv1beta1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	})

	return rules
}
