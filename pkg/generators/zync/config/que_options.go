package config

import (
	saasv1alpha1 "github.com/3scale/saas-operator/api/v1alpha1"
	"github.com/3scale/saas-operator/pkg/generators/common_blocks/pod"
	"k8s.io/utils/pointer"
)

// QueOptions holds configuration for the que pods
type QueOptions struct {
	RailsEnvironment        pod.EnvVarValue `env:"RAILS_ENV"`
	RailsLogLevel           pod.EnvVarValue `env:"RAILS_LOG_LEVEL"`
	RailsLogToStdOut        pod.EnvVarValue `env:"RAILS_LOG_TO_STDOUT"`
	DatabaseURL             pod.EnvVarValue `env:"DATABASE_URL" secret:"zync"`
	SecretKeyBase           pod.EnvVarValue `env:"SECRET_KEY_BASE" secret:"zync"`
	ZyncAuthenticationToken pod.EnvVarValue `env:"ZYNC_AUTHENTICATION_TOKEN" secret:"zync"`
	BugsnagAPIKey           pod.EnvVarValue `env:"BUGSNAG_API_KEY" secret:"zync"`
}

// NewQueOptions returns an Options struct for the given saasv1alpha1.ZyncSpec
func NewQueOptions(spec saasv1alpha1.ZyncSpec) QueOptions {
	opts := QueOptions{
		RailsEnvironment: &pod.ClearTextValue{Value: *spec.Config.Rails.Environment},
		RailsLogLevel:    &pod.ClearTextValue{Value: *spec.Config.Rails.LogLevel},
		RailsLogToStdOut: &pod.ClearTextValue{Value: "true"},

		DatabaseURL:             &pod.SecretValue{Value: spec.Config.DatabaseDSN},
		ZyncAuthenticationToken: &pod.SecretValue{Value: spec.Config.ZyncAuthToken},
	}

	if spec.Config.Bugsnag.Enabled() {
		opts.BugsnagAPIKey = &pod.SecretValue{Value: spec.Config.Bugsnag.APIKey}
	} else {
		opts.BugsnagAPIKey = &pod.SecretValue{Value: saasv1alpha1.SecretReference{Override: pointer.StringPtr("")}}
	}

	return opts
}
