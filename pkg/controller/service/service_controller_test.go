package service

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/kubelix/deployer/pkg/apis/apps/v1alpha1"
)

func TestReconcileService_makeKubelixLabels(t *testing.T) {
	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		svc *appsv1alpha1.Service
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "testing",
			args: args{
				svc: &appsv1alpha1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "testing",
					},
				},
			},
			want: map[string]string{
				"apps.kubelix.io/service": "test",
				"apps.kubelix.io/project": "testing",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileService{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}
			if got := r.makeKubelixLabels(tt.args.svc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeKubelixLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileService_makeLabels(t *testing.T) {
	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}
	type args struct {
		svc *appsv1alpha1.Service
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]string
	}{
		{
			name: "testing",
			args: args{
				svc: &appsv1alpha1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "testing",
					},
				},
			},
			want: map[string]string{
				"apps.kubelix.io/service":      "test",
				"apps.kubelix.io/project":      "testing",
				"app.kubernetes.io/name":       "testing",
				"app.kubernetes.io/svc":        "test",
				"app.kubernetes.io/managed-by": "kubelix-deployer",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileService{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}
			if got := r.makeLabels(tt.args.svc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeKubelixLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
