package mappingservice

import (
	"github.com/3scale-ops/saas-operator/pkg/resource_builders/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (gen *Generator) service() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: gen.GetComponent(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: service.Ports(
				service.TCPPort("mapping", 80, intstr.FromString("mapping")),
				service.TCPPort("management", 8090, intstr.FromString("management")),
			),
		},
	}
}
