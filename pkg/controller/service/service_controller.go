package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Service
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Service{},
	})
	if err != nil {
		return err
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

	// Fetch the Service instance
	instance := &appsv1alpha1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	err = r.ensureDeployment(err, instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileService) ensureDeployment(err error, instance *appsv1alpha1.Service, reqLogger logr.Logger) error {
	// Define a new Deployment object
	dep, err := r.newDeploymentForService(instance)
	if err != nil {
		return err
	}

	depName := types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}
	if err := r.ensureObject(reqLogger, dep, depName); err != nil {
		return fmt.Errorf("failed to handle deployment: %v", err)
	}

	return nil
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
		return fmt.Errorf("failed to update object: %v", err)
	}
	return nil
}

// newDeploymentForService creates the deployment for the given service
func (r *ReconcileService) newDeploymentForService(svc *appsv1alpha1.Service) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"kubelix.io/service": svc.Name,
		"kubelix.io/project": svc.Spec.ProjectName,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:      svc.Name,
							Image:     svc.Spec.Image,
							Command:   svc.Spec.Command,
							Args:      svc.Spec.Args,
							Env:       svc.Spec.Env.ToEnvVars(),
							Resources: svc.Spec.Resources,
							Ports:     svc.Spec.Ports.ToPodPorts(),
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

	// Set Service instance as the owner and controller
	if err := controllerutil.SetControllerReference(svc, dep, r.scheme); err != nil {
		return nil, err
	}

	return dep, nil
}
