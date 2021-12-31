package redisshard

import (
	basereconciler "github.com/3scale/saas-operator/pkg/reconcilers/basereconciler/v1"
	"github.com/MakeNowJust/heredoc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RedisConfigConfigMap returns a basereconciler.GeneratorFunction function that will return a ConfigMap
// resource when called
func (gen *Generator) RedisConfigConfigMap() basereconciler.GeneratorFunction {
	return func() client.Object {
		return &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-config-" + gen.GetInstanceName(),
				Namespace: gen.GetNamespace(),
				Labels:    gen.GetLabels(),
			},
			Data: map[string]string{
				"redis.conf": heredoc.Doc(`
					slaveof 127.0.0.1 6379
					tcp-keepalive 60
					save 900 1
					save 300 10
				`),
			},
		}
	}
}

// RedisReadinessScriptConfigMap returns a basereconciler.GeneratorFunction function that will return a ConfigMap
// resource when called
func (gen *Generator) RedisReadinessScriptConfigMap() basereconciler.GeneratorFunction {
	return func() client.Object {
		return &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-readiness-script-" + gen.GetInstanceName(),
				Namespace: gen.GetNamespace(),
				Labels:    gen.GetLabels(),
			},
			Data: map[string]string{
				"ready.sh": heredoc.Doc(`

					check_master(){
							exit 0
					}

					check_slave(){
							in_sync=$(redis-cli info replication | grep master_sync_in_progress:1 | tr -d "\r" | tr -d "\n")
							no_master=$(redis-cli info replication | grep master_host:127.0.0.1 | tr -d "\r" | tr -d "\n")
							if [ -z "$in_sync" ] && [ -z "$no_master" ]; then
									exit 0
							fi
							exit 1
					}

					role=$(redis-cli info replication | grep role | tr -d "\r" | tr -d "\n")

					case $role in
							role:master)
									check_master
									;;
							role:slave)
									check_slave
									;;
							*)
									echo "unexpected"
									exit 1
					esac
				`),
			},
		}
	}
}
