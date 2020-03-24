package arguments

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

var (
	Kubeconfig *string
	Namespace  *string
	Version    *string
)

// Parses the CLI test arguments and provides access to the stored values.
// Should be set in the TestMain method of the tests to parse CLI Arguments.
//
// Kubeconfig - path to kubeconfig file. If sets it do the default value
//
// Namespace - Namespace for deployment. If not provided sets it to a generated namespace
//
// Version - tag to be used to find the correct docker image
func Parse() {
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	Namespace = flag.String("namespace", "", "Default namespace")
	Version = flag.String("version", "master", "Pod version")
	flag.Parse()
}
