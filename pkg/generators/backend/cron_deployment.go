package backend

import (
	"fmt"
	"strings"

	"github.com/3scale/saas-operator/pkg/generators/common_blocks/pod"
	"github.com/3scale/saas-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment returns a function that will return a Deployment
// resource when called
func (gen *CronGenerator) deployment() func() *appsv1.Deployment {

	return func() *appsv1.Deployment {

		dep := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Replicas: gen.CronSpec.Replicas,
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: util.IntStrPtr(intstr.FromInt(0)),
						MaxSurge:       util.IntStrPtr(intstr.FromInt(1)),
					},
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						ImagePullSecrets: func() []corev1.LocalObjectReference {
							if gen.Image.PullSecretName != nil {
								return []corev1.LocalObjectReference{{Name: *gen.Image.PullSecretName}}
							}
							return nil
						}(),
						Containers: []corev1.Container{
							{
								Name:                     strings.Join([]string{component, cron}, "-"),
								Image:                    fmt.Sprintf("%s:%s", *gen.Image.Name, *gen.Image.Tag),
								Args:                     []string{"backend-cron"},
								Env:                      pod.BuildEnvironment(gen.Options),
								Resources:                corev1.ResourceRequirements(*gen.CronSpec.Resources),
								ImagePullPolicy:          *gen.Image.PullPolicy,
								TerminationMessagePath:   corev1.TerminationMessagePathDefault,
								TerminationMessagePolicy: corev1.TerminationMessageReadFile,
							},
						},
						Affinity:    pod.Affinity(gen.GetSelector(), gen.CronSpec.NodeAffinity),
						Tolerations: gen.CronSpec.Tolerations,
					},
				},
			},
		}
		return dep
	}
}
