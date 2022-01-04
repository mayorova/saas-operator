package backend

import (
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// service returns a function that will return the
// public service resource when called
func (gen *ListenerGenerator) service() func() *corev1.Service {

	return func() *corev1.Service {

		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        gen.GetComponent(),
				Annotations: service.NLBServiceAnnotations(*gen.ListenerSpec.LoadBalancer, gen.ListenerSpec.Endpoint.DNS),
			},
			Spec: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeCluster,
				SessionAffinity:       corev1.ServiceAffinityNone,
				Ports: func() []corev1.ServicePort {
					if gen.ListenerSpec.Marin3r.IsDeactivated() {
						return service.Ports(
							service.TCPPort("http", 80, intstr.FromString("http")),
						)
					}
					return service.Ports(
						service.TCPPort("http", 80, intstr.FromString("backend-http")),
						service.TCPPort("https", 443, intstr.FromString("backend-https")),
					)
				}(),
			},
		}
	}
}

// internalService returns a function that will return the
// internal Service resource when called
func (gen *ListenerGenerator) internalService() func() *corev1.Service {

	return func() *corev1.Service {

		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: gen.GetComponent() + "-internal",
			},
			Spec: corev1.ServiceSpec{
				Type:            corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
				Ports: func() []corev1.ServicePort {
					if gen.ListenerSpec.Marin3r.IsDeactivated() {
						return service.Ports(
							service.TCPPort("http", 80, intstr.FromString("http")),
						)
					}
					return service.Ports(
						service.TCPPort("http", 80, intstr.FromString("http-internal")),
					)
				}(),
			},
		}
	}
}
