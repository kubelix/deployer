package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/kubelix/deployer/pkg/apis/apps/v1alpha1"
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

	secrets, err := r.ensureDockerPullSecrets(svc, reqLogger.WithValues("Generated.Version", "core/v1", "Generated.Kind", "Secret"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureFilesConfigMap(svc, reqLogger.WithValues("Generated.Version", "core/v1", "Generated.Kind", "ConfigMap"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureDeployment(svc, secrets, reqLogger.WithValues("Generated.Version", "apps/v1", "Generated.Kind", "Deployment"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureService(svc, reqLogger.WithValues("Generated.Version", "core/v1", "Generated.Kind", "Service"))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureIngresses(svc, reqLogger.WithValues("Generated.Version", "networking/v1beta1", "Generated.Kind", "Ingress"))
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileService) makeKubelixLabels(svc *appsv1alpha1.Service) map[string]string {
	return map[string]string{
		"apps.kubelix.io/service": svc.Name,
		"apps.kubelix.io/project": svc.Namespace,
	}
}

func (r *ReconcileService) makeLabels(svc *appsv1alpha1.Service) map[string]string {
	return mergeLabels(r.makeKubelixLabels(svc), map[string]string{
		"app.kubernetes.io/name":       svc.Namespace,
		"app.kubernetes.io/svc":        svc.Name,
		"app.kubernetes.io/managed-by": "kubelix-deployer",
	})
}

func (r *ReconcileService) ensureObject(reqLogger logr.Logger, svc *appsv1alpha1.Service, obj runtime.Object, name types.NamespacedName) error {
	found := obj.DeepCopyObject()
	err := r.client.Get(context.TODO(), name, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating object", "Type.Namespace", name.Namespace, "Type.Name", name.Name)
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			reqLogger.Error(err, fmt.Sprintf("Object: %#v", obj))
			return fmt.Errorf("failed to create object: %v", err)
		}
		return nil
	} else if err != nil {
		return err
	}

	kind := obj.GetObjectKind().GroupVersionKind().String()
	sumKey := fmt.Sprintf("%s/%s/%s", kind, name.Namespace, name.Name)
	sum, err := Checksum(obj)
	if err != nil {
		return fmt.Errorf("failed to get checksum of deployment body: %v", err)
	}

	if oldSum, ok := svc.Status.Checksums[sumKey]; ok && oldSum == sum {
		reqLogger.Info("Checksums of old and new object match, do not update", "Type.Namespace", name.Namespace, "Type.Name", name.Name)
		return nil
	}
	svc.Status.Checksums[sumKey] = sum

	reqLogger.Info("Updating existing object", "Type.Namespace", name.Namespace, "Type.Name", name.Name)

	err = r.client.Update(context.TODO(), obj)
	if err != nil {
		if strings.Contains(err.Error(), "field is immutable") {
			errDelete := r.client.Delete(context.TODO(), obj)
			if errDelete != nil {
				return fmt.Errorf("failed to delete object after update was not permitted (field is immutable): %v", errDelete)
			}

			// as we have deleted the object we now can safely recreate it
			return r.ensureObject(reqLogger, svc, obj, name)
		}
		return fmt.Errorf("failed to update object: %v", err)
	}
	return nil
}
