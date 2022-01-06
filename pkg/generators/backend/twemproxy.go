package backend

import (
	"path/filepath"

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/generators/backend/config"
	"github.com/3scale/saas-operator/pkg/resource_builders/pod"
	"github.com/3scale/saas-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	twemproxyPreStopScriptName = "pre-stop"
	healthCommand              = "health"
)

func addTwemproxySidecar(dep appsv1.Deployment, spec *saasv1alpha1.TwemproxySpec) *appsv1.Deployment {

	// Labels to subscribe to the TwemproxyConfig sync events
	dep.Spec.Template.ObjectMeta.Labels = util.MergeMaps(
		map[string]string{},
		dep.Spec.Template.GetLabels(),
		map[string]string{saasv1alpha1.TwemproxyPodSyncLabelKey: spec.TwemproxyConfigRef},
	)

	// Twemproxy container
	dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers,
		corev1.Container{
			Env:   pod.BuildEnvironment(config.NewTwemproxyOptions(*spec)),
			Name:  twemproxy,
			Image: pod.Image(*spec.Image),
			Ports: pod.ContainerPorts(
				pod.ContainerPortTCP(twemproxy, 22121),
				pod.ContainerPortTCP("twem-metrics", int32(*spec.Options.MetricsPort)),
			),
			Resources:                corev1.ResourceRequirements(*spec.Resources),
			ImagePullPolicy:          *spec.Image.PullPolicy,
			LivenessProbe:            pod.ExecProbe(healthCommand, *spec.LivenessProbe),
			ReadinessProbe:           pod.ExecProbe(healthCommand, *spec.ReadinessProbe),
			TerminationMessagePath:   corev1.TerminationMessagePathDefault,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{"pre-stop", config.TwemproxyConfigFile},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      twemproxy + "-config",
					MountPath: filepath.Dir(config.TwemproxyConfigFile),
				},
			},
		})

	if dep.Spec.Template.Spec.Volumes == nil {
		dep.Spec.Template.Spec.Volumes = []corev1.Volume{}
	}

	// Mount the TwemproxyConfig ConfigMap in the Pod
	dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: twemproxy + "-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: spec.ConfigMapName(),
					},
					DefaultMode: pointer.Int32(420),
				},
			},
		})

	return &dep
}
