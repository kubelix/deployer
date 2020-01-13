package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "gitlab.com/klinkert.io/kubelix/deployer/pkg/apis/apps/v1alpha1"
)

var log = logf.Log.WithName("controller_service")

// Add creates a new Service Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	createdItems := []runtime.Object{
		//&appsv1.Deployment{},
		//&corev1.Service{},
		//&networkingv1beta1.Ingress{},
	}

	for _, t := range createdItems {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &appsv1alpha1.Service{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a Service object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Service.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Service")

	// Fetch the Service svc
	svc := &appsv1alpha1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = r.ensureDeployment(err, svc, reqLogger.WithValues("Generated.Version", "apps/v1", "Generated.Kind", "Deployment"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureService(err, svc, reqLogger.WithValues("Generated.Version", "core/v1", "Generated.Kind", "Service"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureIngresses(err, svc, reqLogger.WithValues("Generated.Version", "networking/v1beta1", "Generated.Kind", "Ingress"))
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileService) makeLabels(svc *appsv1alpha1.Service) map[string]string {
	return map[string]string{
		"apps.kubelix.io/service": svc.Name,
		"apps.kubelix.io/project": svc.Spec.ProjectName,

		"app.kubernetes.io/name":       svc.Spec.ProjectName,
		"app.kubernetes.io/svc":        svc.Name,
		"app.kubernetes.io/managed-by": "kubelix-deployer",
	}
}

func (r *ReconcileService) ensureObject(reqLogger logr.Logger, obj runtime.Object, name types.NamespacedName) error {
	found := obj.DeepCopyObject()
	err := r.client.Get(context.TODO(), name, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating object", "Pod.Namespace", name.Namespace, "Pod.Name", name.Name)
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			return fmt.Errorf("failed to create object: %v", err)
		}
		return nil
	} else if err != nil {
		return err
	}

	reqLogger.Info("Updating existing object", "Pod.Namespace", name.Namespace, "Pod.Name", name.Name)

	err = r.client.Update(context.TODO(), obj)
	if err != nil {
		if strings.Contains(err.Error(), "field is immutable") {
			errDelete := r.client.Delete(context.TODO(), obj)
			if errDelete != nil {
				return fmt.Errorf("failed to delete object after update was not permitted (field is immutable): %v", errDelete)
			}

			// as we have deleted the object we now can safely recreate it
			return r.ensureObject(reqLogger, obj, name)
		}
		return fmt.Errorf("failed to update object: %v", err)
	}
	return nil
}

func (r *ReconcileService) ensureDeployment(err error, svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
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

// newDeploymentForService creates the deployment for the given service
func (r *ReconcileService) newDeploymentForService(svc *appsv1alpha1.Service) (*appsv1.Deployment, error) {
	labels := r.makeLabels(svc)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
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

	if err := controllerutil.SetControllerReference(svc, dep, r.scheme); err != nil {
		return nil, err
	}

	return dep, nil
}

func (r *ReconcileService) ensureService(err error, svc *appsv1alpha1.Service, reqLogger logr.Logger) error {
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

// newDeploymentForService creates the deployment for the given service
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

	if err := controllerutil.SetControllerReference(svc, coreService, r.scheme); err != nil {
		return nil, err
	}

	return coreService, nil
}

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

// newDeploymentForService creates the deployment for the given coreService
func (r *ReconcileService) newIngressesForService(svc *appsv1alpha1.Service) ([]*networkingv1beta1.Ingress, error) {
	labels := r.makeLabels(svc)
	ingresses := make([]*networkingv1beta1.Ingress, 0)

	for _, p := range svc.Spec.Ports {
		if len(p.Ingresses) == 0 {
			continue
		}

		for _, ing := range p.Ingresses {
			ingress := &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svc.Name,
					Namespace: svc.Namespace,
					Labels:    labels,
				},
				Spec: networkingv1beta1.IngressSpec{
					Rules: r.makeIngressRules(svc, p, ing),
				},
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
