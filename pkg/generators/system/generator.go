package system

import (
	"strings"

	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/basereconciler"
	"github.com/3scale/saas-operator/pkg/generators"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/grafanadashboard"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/hpa"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/pdb"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/pod"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/podmonitor"
	"github.com/3scale/saas-operator/pkg/generators/system/config"
	"k8s.io/apimachinery/pkg/types"
)

const (
	component string = "system"
	app       string = "app"
	sidekiq   string = "sidekiq"
	sphinx    string = "sphinx"

	systemConfigSecret = "system-config"
)

// Generator configures the generators for Apicast
type Generator struct {
	generators.BaseOptions
	App                  AppGenerator
	Sidekiq              SidekiqGenerator
	Sphinx               SphinxGenerator
	GrafanaDashboardSpec saasv1alpha1.GrafanaDashboardSpec
	ConfigFilesSpec      saasv1alpha1.ConfigFilesSpec
	Options              config.Options
}

// Dashboard returns a basereconciler.GeneratorFunction
func (gen *Generator) Dashboard() basereconciler.GeneratorFunction {
	key := types.NamespacedName{Name: gen.Component, Namespace: gen.Namespace}
	return grafanadashboard.New(key, gen.GetLabels(), gen.GrafanaDashboardSpec, "dashboards/system.json.tpl")
}

// DatabaseSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) DatabaseSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-database", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// SeedSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) SeedSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-seed", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// RecaptchaSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) RecaptchaSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-recaptcha", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// EventsHookSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) EventsHookSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-events-hook", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// SMTPSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) SMTPSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-smtp", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// MasterApicastSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) MasterApicastSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-master-apicast", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// ZyncSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) ZyncSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-zync", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// BackendSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) BackendSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-backend", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// MultitenantAssetsSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) MultitenantAssetsSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-multitenant-assets-s3", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// AppSecretDefinition returns a basereconciler.GeneratorFunction
func (gen *Generator) AppSecretDefinition() basereconciler.GeneratorFunction {
	return pod.GenerateSecretDefinitionFn("system-app", gen.GetNamespace(), gen.GetLabels(), gen.Options)
}

// NewGenerator returns a new Options struct
func NewGenerator(instance, namespace string, spec saasv1alpha1.SystemSpec) Generator {
	return Generator{
		BaseOptions: generators.BaseOptions{
			Component:    component,
			InstanceName: instance,
			Namespace:    namespace,
			Labels: map[string]string{
				"app":                  "3scale-api-management",
				"threescale_component": component,
			},
		},
		App: AppGenerator{
			BaseOptions: generators.BaseOptions{
				Component:    strings.Join([]string{component, app}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": app,
				},
			},
			Spec:      *spec.App,
			Options:   config.NewOptions(spec),
			ImageSpec: *spec.Image,
		},
		Sidekiq: SidekiqGenerator{
			BaseOptions: generators.BaseOptions{
				Component:    strings.Join([]string{component, sidekiq}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": sidekiq,
				},
			},
			Spec:      *spec.Sidekiq,
			Options:   config.NewOptions(spec),
			ImageSpec: *spec.Image,
		},
		Sphinx: SphinxGenerator{
			BaseOptions: generators.BaseOptions{
				Component:    strings.Join([]string{component, sphinx}, "-"),
				InstanceName: instance,
				Namespace:    namespace,
				Labels: map[string]string{
					"app":                          "3scale-api-management",
					"threescale_component":         component,
					"threescale_component_element": sphinx,
				},
			},
			Spec:      *spec.Sphinx,
			Config:    spec.Config,
			ImageSpec: *spec.Image,
		},
		GrafanaDashboardSpec: *spec.GrafanaDashboard,
		ConfigFilesSpec:      spec.Config.ConfigFiles,
		Options:              config.NewOptions(spec),
	}
}

// AppGenerator has methods to generate resources for system-app
type AppGenerator struct {
	generators.BaseOptions
	Spec      saasv1alpha1.SystemAppSpec
	Options   config.Options
	ImageSpec saasv1alpha1.ImageSpec
}

// HPA returns a basereconciler.GeneratorFunction
func (gen *AppGenerator) HPA() basereconciler.GeneratorFunction {
	return hpa.New(gen.Key(), gen.GetLabels(), *gen.Spec.HPA)
}

// PDB returns a basereconciler.GeneratorFunction
func (gen *AppGenerator) PDB() basereconciler.GeneratorFunction {
	return pdb.New(gen.Key(), gen.GetLabels(), gen.Selector().MatchLabels, *gen.Spec.PDB)
}

// PodMonitor returns a basereconciler.GeneratorFunction
func (gen *AppGenerator) PodMonitor() basereconciler.GeneratorFunction {
	return podmonitor.New(gen.Key(), gen.GetLabels(), gen.Selector().MatchLabels,
		podmonitor.PodMetricsEndpoint("/metrics", "metrics", 30),
		podmonitor.PodMetricsEndpoint("/yabeda-metrics", "metrics", 30),
	)
}

// SidekiqGenerator has methods to generate resources for system-sidekiq
type SidekiqGenerator struct {
	generators.BaseOptions
	Spec      saasv1alpha1.SystemSidekiqSpec
	Options   config.Options
	ImageSpec saasv1alpha1.ImageSpec
}

// HPA returns a basereconciler.GeneratorFunction
func (gen *SidekiqGenerator) HPA() basereconciler.GeneratorFunction {
	key := types.NamespacedName{Name: gen.Component, Namespace: gen.Namespace}
	return hpa.New(key, gen.GetLabels(), *gen.Spec.HPA)
}

// PDB returns a basereconciler.GeneratorFunction
func (gen *SidekiqGenerator) PDB() basereconciler.GeneratorFunction {
	key := types.NamespacedName{Name: gen.Component, Namespace: gen.Namespace}
	return pdb.New(key, gen.GetLabels(), gen.Selector().MatchLabels, *gen.Spec.PDB)
}

// PodMonitor returns a basereconciler.GeneratorFunction
func (gen *SidekiqGenerator) PodMonitor() basereconciler.GeneratorFunction {
	key := types.NamespacedName{Name: gen.Component, Namespace: gen.Namespace}
	return podmonitor.New(key, gen.GetLabels(), gen.Selector().MatchLabels,
		podmonitor.PodMetricsEndpoint("/metrics", "metrics", 30),
	)
}

// SphinxGenerator has methods to generate resources for system-sphinx
type SphinxGenerator struct {
	generators.BaseOptions
	Spec      saasv1alpha1.SystemSphinxSpec
	Config    saasv1alpha1.SystemConfig
	ImageSpec saasv1alpha1.ImageSpec
}
