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

	//createdItems := []runtime.Object{
	//	//&appsv1.Deployment{},
	//	//&corev1.Service{},
	//	//&networkingv1beta1.Ingress{},
	//}
	//
	//for _, t := range createdItems {
	//	err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
	//		IsController: true,
	//		OwnerType:    &appsv1alpha1.Service{},
	//	})
	//	if err != nil {
	//		return err
	//	}
	//}

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

	// Fetch the Service object
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

	generatedObjects := make([]runtime.Object, 0)

	secrets, err := r.ensureDockerPullSecrets(svc, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}
	for _, s := range secrets {
		generatedObjects = append(generatedObjects, s)
	}

	configMap, err := r.ensureFilesConfigMap(svc, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}
	generatedObjects = append(generatedObjects, configMap)

	dep, err := r.ensureDeployment(svc, secrets, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}
	generatedObjects = append(generatedObjects, dep)

	if len(svc.Spec.Ports) > 0 {
		coreService, err := r.ensureService(svc, reqLogger)
		if err != nil {
			return reconcile.Result{}, err
		}
		generatedObjects = append(generatedObjects, coreService)

		ingresses, err := r.ensureIngresses(svc, reqLogger)
		if err != nil {
			return reconcile.Result{}, err
		}
		for _, i := range ingresses {
			generatedObjects = append(generatedObjects, i)
		}
	}

	if err := r.cleanupManagedObjects(reqLogger, svc, generatedObjects); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, r.update(reqLogger, svc)
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
	objGVK := obj.GetObjectKind().GroupVersionKind()
	reqLogger = reqLogger.WithValues(
		"Generated.Version", objGVK.GroupVersion(),
		"Generated.Kind", objGVK.GroupKind().Kind,
		"Type.Namespace", name.Namespace,
		"Type.Name", name.Name,
	)

	defer func() {
		obj.GetObjectKind().SetGroupVersionKind(objGVK)
	}()

	err, match := r.setManagedObject(reqLogger, svc, obj, name)
	if err != nil {
		return err
	} else if match {
		reqLogger.Info("Checksums of old and new object match, do not update")
		return nil
	}

	found := obj.DeepCopyObject()
	err = r.client.Get(context.TODO(), name, found)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Creating object")
			err = r.client.Create(context.TODO(), obj)
			if err != nil {
				reqLogger.Error(err, fmt.Sprintf("Object: %#v", obj))
				return fmt.Errorf("failed to create object: %v", err)
			}

			return r.update(reqLogger, svc)
		}

		return err
	}

	reqLogger.Info("Updating existing object")

	err = r.client.Update(context.TODO(), obj)
	if err != nil {
		if strings.Contains(err.Error(), fieldIsImmutable) {
			errDelete := r.client.Delete(context.TODO(), obj)
			if errDelete != nil {
				return fmt.Errorf("failed to delete object after update was not permitted (field is immutable): %v", errDelete)
			}

			// as we have deleted the object we now can safely recreate it
			return r.ensureObject(reqLogger, svc, obj, name)
		}
		return fmt.Errorf("failed to update object: %v", err)
	}

	return r.update(reqLogger, svc)
}

func (r *ReconcileService) update(reqLogger logr.Logger, svc *appsv1alpha1.Service) error {
	if err := r.client.Status().Update(context.TODO(), svc); err != nil {
		return fmt.Errorf("failed to update status: %v", err)
	}

	if err := r.client.Update(context.TODO(), svc); err != nil {
		return fmt.Errorf("failed to update: %v", err)
	}

	name := types.NamespacedName{Namespace: svc.Namespace, Name: svc.Name}
	if err := r.client.Get(context.TODO(), name, svc); err != nil {
		return fmt.Errorf("failed to reload service from API: %v", err)
	}

	return nil
}

func (r *ReconcileService) setManagedObject(reqLogger logr.Logger, svc *appsv1alpha1.Service, obj runtime.Object, name types.NamespacedName) (error, bool) {
	checksum, err := checksum(obj)
	if err != nil {
		return fmt.Errorf("failed to get checksum of object (%s %s): %v", obj, name, err), false
	}

	managedObject := svc.Status.ManagedObjects.Find(obj, name)
	if managedObject == nil {
		svc.Status.ManagedObjects.Add(obj, name, checksum)
		return nil, false
	}

	if managedObject.Checksum == checksum {
		return nil, true
	}

	return nil, false
}

func (r *ReconcileService) cleanupManagedObjects(reqLogger logr.Logger, svc *appsv1alpha1.Service, generatedObjects []runtime.Object) error {
	newList := appsv1alpha1.ManagedObjectList{}
	newList.FromObjectList(generatedObjects)

	for _, ref := range svc.Status.ManagedObjects {
		// managedObject is also contained by current version, so the object was not deleted
		if newList.Contains(ref) {
			continue
		}

		if err := r.deleteManagedObject(reqLogger, ref); err != nil {
			return fmt.Errorf("failed to clean up object: %v", err)
		}

		svc.Status.ManagedObjects.Remove(ref)
	}

	return nil
}

func (r *ReconcileService) deleteManagedObject(reqLogger logr.Logger, managedObject *appsv1alpha1.ManagedObject) error {
	kind := managedObject.GroupVersionKind()
	name := managedObject.NamespacedName()

	obj, err := r.scheme.New(kind)
	if err != nil {
		return fmt.Errorf("failed to create object from managedObject %s: %v", managedObject, err)
	}

	err = r.client.Get(context.TODO(), name, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Error(err, fmt.Sprintf("failed to delete object %s", managedObject))
			return nil
		}

		return fmt.Errorf("failed to find managedObject %s: %v", managedObject, err)
	}

	err = r.client.Delete(context.TODO(), obj)
	if err != nil {
		return fmt.Errorf("failed to delete managedObject %s: %v", managedObject, err)
	}

	reqLogger.Info(fmt.Sprintf("deleted managedObject %s", managedObject))

	return nil
}
