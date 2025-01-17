package backend

import (
	"fmt"
	"strings"

	"github.com/3scale-ops/basereconciler/mutators"
	"github.com/3scale-ops/basereconciler/resource"
	"github.com/3scale-ops/basereconciler/util"
	saasv1alpha1 "github.com/3scale-ops/saas-operator/api/v1alpha1"
	"github.com/3scale-ops/saas-operator/pkg/generators"
	"github.com/3scale-ops/saas-operator/pkg/generators/backend/config"
	descriptor "github.com/3scale-ops/saas-operator/pkg/resource_builders/envoyconfig/descriptor"
	"github.com/3scale-ops/saas-operator/pkg/resource_builders/grafanadashboard"
	"github.com/3scale-ops/saas-operator/pkg/resource_builders/pod"
	"github.com/3scale-ops/saas-operator/pkg/resource_builders/podmonitor"
	operatorutil "github.com/3scale-ops/saas-operator/pkg/util"
	deployment_workload "github.com/3scale-ops/saas-operator/pkg/workloads/deployment"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	component string = "backend"
	listener  string = "listener"
	worker    string = "worker"
	cron      string = "cron"
)

// Generator configures the generators for Backend
type Generator struct {
	generators.BaseOptionsV2
	Listener             ListenerGenerator
	CanaryListener       *ListenerGenerator
	Worker               WorkerGenerator
	CanaryWorker         *WorkerGenerator
	Cron                 CronGenerator
	grafanaDashboardSpec saasv1alpha1.GrafanaDashboardSpec
	config               saasv1alpha1.BackendConfig
}

// NewGenerator returns a new Options struct
func NewGenerator(instance, namespace string, spec saasv1alpha1.BackendSpec) (Generator, error) {

	generator := Generator{
		BaseOptionsV2: generators.BaseOptionsV2{
			Component:    component,
			InstanceName: instance,
			Namespace:    namespace,
			Labels: map[string]string{
				"app":                  "3scale-api-management",
				"threescale_component": component,
			},
		},
		Listener: ListenerGenerator{
			BaseOptionsV2: generators.BaseOptionsV2{
				Component:    strings.Join([]string{component, listener}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": listener,
				},
			},
			ListenerSpec:  spec.Listener,
			Image:         *spec.Image,
			Options:       config.NewListenerOptions(spec),
			Traffic:       true,
			TwemproxySpec: spec.Twemproxy,
		},
		Worker: WorkerGenerator{
			BaseOptionsV2: generators.BaseOptionsV2{
				Component:    strings.Join([]string{component, worker}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": worker,
				},
			},
			WorkerSpec:    *spec.Worker,
			Image:         *spec.Image,
			Options:       config.NewWorkerOptions(spec),
			TwemproxySpec: spec.Twemproxy,
		},
		Cron: CronGenerator{
			BaseOptionsV2: generators.BaseOptionsV2{
				Component:    strings.Join([]string{component, cron}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": cron,
				},
			},
			CronSpec: *spec.Cron,
			Image:    *spec.Image,
			Options:  config.NewCronOptions(spec),
		},
		grafanaDashboardSpec: *spec.GrafanaDashboard,
		config:               spec.Config,
	}

	if spec.Listener.Canary != nil {
		canarySpec, err := spec.ResolveCanarySpec(spec.Listener.Canary)
		if err != nil {
			return Generator{}, err
		}
		generator.CanaryListener = &ListenerGenerator{
			BaseOptionsV2: generators.BaseOptionsV2{
				Component:    strings.Join([]string{component, listener, "canary"}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": "canary-" + listener,
				},
			},
			ListenerSpec:  canarySpec.Listener,
			Image:         *canarySpec.Image,
			Options:       config.NewListenerOptions(*canarySpec),
			Traffic:       spec.Listener.Canary.SendTraffic,
			TwemproxySpec: canarySpec.Twemproxy,
		}
		// Disable PDB and HPA for the canary Deployment
		generator.CanaryListener.ListenerSpec.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
		generator.CanaryListener.ListenerSpec.PDB = &saasv1alpha1.PodDisruptionBudgetSpec{}
	}

	if spec.Worker.Canary != nil {
		canarySpec, err := spec.ResolveCanarySpec(spec.Worker.Canary)
		if err != nil {
			return Generator{}, err
		}
		generator.CanaryWorker = &WorkerGenerator{
			BaseOptionsV2: generators.BaseOptionsV2{
				Component:    strings.Join([]string{component, worker, "canary"}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": "canary-" + worker,
				},
			},
			WorkerSpec:    *canarySpec.Worker,
			Image:         *canarySpec.Image,
			Options:       config.NewWorkerOptions(*canarySpec),
			TwemproxySpec: canarySpec.Twemproxy,
		}
		// Disable PDB and HPA for the canary Deployment
		generator.CanaryWorker.WorkerSpec.HPA = &saasv1alpha1.HorizontalPodAutoscalerSpec{}
		generator.CanaryWorker.WorkerSpec.PDB = &saasv1alpha1.PodDisruptionBudgetSpec{}
	}

	return generator, nil
}

// Resources returns the list of resource templates
func (gen *Generator) Resources() ([]resource.TemplateInterface, error) {
	listener_resources, err := deployment_workload.New(&gen.Listener, gen.CanaryListener)
	if err != nil {
		return nil, err
	}
	worker_resources, err := deployment_workload.New(&gen.Worker, gen.CanaryWorker)
	if err != nil {
		return nil, err
	}
	cron_resources, err := deployment_workload.New(&gen.Cron, nil)
	if err != nil {
		return nil, err
	}

	misc := []resource.TemplateInterface{
		// GrafanaDashboard
		resource.NewTemplate(
			grafanadashboard.New(gen.GetKey(), gen.GetLabels(), gen.grafanaDashboardSpec, "dashboards/backend.json.gtpl")).
			WithEnabled(!gen.grafanaDashboardSpec.IsDeactivated()),
		// ExternalSecrets
		resource.NewTemplate(
			pod.GenerateExternalSecretFn("backend-system-events-hook", gen.GetNamespace(),
				*gen.config.ExternalSecret.SecretStoreRef.Name, *gen.config.ExternalSecret.SecretStoreRef.Kind,
				*gen.config.ExternalSecret.RefreshInterval, gen.GetLabels(), gen.Worker.Options),
		),
		resource.NewTemplate(
			pod.GenerateExternalSecretFn("backend-internal-api", gen.GetNamespace(),
				*gen.config.ExternalSecret.SecretStoreRef.Name, *gen.config.ExternalSecret.SecretStoreRef.Kind,
				*gen.config.ExternalSecret.RefreshInterval, gen.GetLabels(), gen.Listener.Options),
		),
		resource.NewTemplate(
			pod.GenerateExternalSecretFn("backend-error-monitoring", gen.GetNamespace(),
				*gen.config.ExternalSecret.SecretStoreRef.Name, *gen.config.ExternalSecret.SecretStoreRef.Kind,
				*gen.config.ExternalSecret.RefreshInterval, gen.GetLabels(), gen.Listener.Options)).
			WithEnabled(gen.config.ErrorMonitoringKey != nil),
	}

	return operatorutil.ConcatSlices(listener_resources, worker_resources, cron_resources, misc), nil
}

// ListenerGenerator has methods to generate resources for a
// Backend environment
type ListenerGenerator struct {
	generators.BaseOptionsV2
	Image         saasv1alpha1.ImageSpec
	ListenerSpec  saasv1alpha1.ListenerSpec
	Options       config.ListenerOptions
	Traffic       bool
	TwemproxySpec *saasv1alpha1.TwemproxySpec
}

// Validate that ListenerGenerator implements deployment_workload.DeploymentWorkloadWithTraffic interface
var _ deployment_workload.DeploymentWorkload = &ListenerGenerator{}

// Validate that ListenerGenerator implements deployment_workload.WithTraffic interface
var _ deployment_workload.WithTraffic = &ListenerGenerator{}

// Validate that ListenerGenerator implements deployment_workload.WithEnvoySidecar interface
var _ deployment_workload.WithEnvoySidecar = &ListenerGenerator{}

func (gen *ListenerGenerator) Labels() map[string]string {
	return gen.GetLabels()
}
func (gen *ListenerGenerator) Deployment() *resource.Template[*appsv1.Deployment] {
	return resource.NewTemplateFromObjectFunction(gen.deployment).
		WithMutation(mutators.SetDeploymentReplicas(gen.ListenerSpec.HPA.IsDeactivated())).
		WithMutation(mutators.RolloutTrigger{Name: "backend-internal-api", SecretName: util.Pointer("backend-internal-api")}.Add()).
		WithMutation(mutators.RolloutTrigger{Name: "backend-error-monitoring", SecretName: util.Pointer("backend-error-monitoring")}.Add())
}

func (gen *ListenerGenerator) HPASpec() *saasv1alpha1.HorizontalPodAutoscalerSpec {
	return gen.ListenerSpec.HPA
}
func (gen *ListenerGenerator) PDBSpec() *saasv1alpha1.PodDisruptionBudgetSpec {
	return gen.ListenerSpec.PDB
}
func (gen *ListenerGenerator) MonitoredEndpoints() []monitoringv1.PodMetricsEndpoint {
	pmes := []monitoringv1.PodMetricsEndpoint{
		podmonitor.PodMetricsEndpoint("/metrics", "metrics", 30),
		podmonitor.PodMetricsEndpoint("/stats/prometheus", "envoy-metrics", 60),
	}
	if gen.TwemproxySpec != nil {
		pmes = append(pmes, podmonitor.PodMetricsEndpoint("/metrics", "twem-metrics", 30))
	}
	return pmes
}
func (gen *ListenerGenerator) Services() []*resource.Template[*corev1.Service] {
	return []*resource.Template[*corev1.Service]{
		resource.NewTemplateFromObjectFunction(gen.service).WithMutation(mutators.SetServiceLiveValues()),
		resource.NewTemplateFromObjectFunction(gen.internalService).WithMutation(mutators.SetServiceLiveValues()),
	}
}
func (gen *ListenerGenerator) SendTraffic() bool { return gen.Traffic }
func (gen *ListenerGenerator) TrafficSelector() map[string]string {
	return map[string]string{
		// This is purposedly hardcoded as the TrafficSelector needs to be the same for all workloads produced
		// by the same generator so traffic can be sent to all of them at the same time
		fmt.Sprintf("%s/traffic", saasv1alpha1.GroupVersion.Group): fmt.Sprintf("%s-%s", component, listener),
	}
}
func (gen *ListenerGenerator) EnvoyDynamicConfigurations() []descriptor.EnvoyDynamicConfigDescriptor {
	return gen.ListenerSpec.Marin3r.EnvoyDynamicConfig.AsList()
}

// WorkerGenerator has methods to generate resources for a
// Backend environment
type WorkerGenerator struct {
	generators.BaseOptionsV2
	Image         saasv1alpha1.ImageSpec
	WorkerSpec    saasv1alpha1.WorkerSpec
	Options       config.WorkerOptions
	TwemproxySpec *saasv1alpha1.TwemproxySpec
}

// Validate that WorkerGenerator implements deployment_workload.DeploymentWorkload interface
var _ deployment_workload.DeploymentWorkload = &WorkerGenerator{}

func (gen *WorkerGenerator) Deployment() *resource.Template[*appsv1.Deployment] {
	return resource.NewTemplateFromObjectFunction(gen.deployment).
		WithMutation(mutators.SetDeploymentReplicas(gen.WorkerSpec.HPA.IsDeactivated())).
		WithMutation(mutators.RolloutTrigger{Name: "backend-system-events-hook", SecretName: util.Pointer("backend-system-events-hook")}.Add()).
		WithMutation(mutators.RolloutTrigger{Name: "backend-error-monitoring", SecretName: util.Pointer("backend-error-monitoring")}.Add())
}
func (gen *WorkerGenerator) HPASpec() *saasv1alpha1.HorizontalPodAutoscalerSpec {
	return gen.WorkerSpec.HPA
}
func (gen *WorkerGenerator) PDBSpec() *saasv1alpha1.PodDisruptionBudgetSpec {
	return gen.WorkerSpec.PDB
}
func (gen *WorkerGenerator) MonitoredEndpoints() []monitoringv1.PodMetricsEndpoint {
	pmes := []monitoringv1.PodMetricsEndpoint{
		podmonitor.PodMetricsEndpoint("/metrics", "metrics", 30),
	}
	if gen.TwemproxySpec != nil {
		pmes = append(pmes, podmonitor.PodMetricsEndpoint("/metrics", "twem-metrics", 30))
	}
	return pmes
}

// CronGenerator has methods to generate resources for a
// Backend environment
type CronGenerator struct {
	generators.BaseOptionsV2
	Image    saasv1alpha1.ImageSpec
	CronSpec saasv1alpha1.CronSpec
	Options  config.CronOptions
}

// Validate that CronGenerator implements deployment_workload.DeploymentWorkload interface
var _ deployment_workload.DeploymentWorkload = &CronGenerator{}

func (gen *CronGenerator) Deployment() *resource.Template[*appsv1.Deployment] {
	return resource.NewTemplateFromObjectFunction(gen.deployment).
		WithMutation(mutators.RolloutTrigger{Name: "backend-error-monitoring", SecretName: util.Pointer("backend-error-monitoring")}.Add())
}
func (gen *CronGenerator) HPASpec() *saasv1alpha1.HorizontalPodAutoscalerSpec {
	return &saasv1alpha1.HorizontalPodAutoscalerSpec{}
}
func (gen *CronGenerator) PDBSpec() *saasv1alpha1.PodDisruptionBudgetSpec {
	return &saasv1alpha1.PodDisruptionBudgetSpec{}
}
func (gen *CronGenerator) MonitoredEndpoints() []monitoringv1.PodMetricsEndpoint { return nil }
