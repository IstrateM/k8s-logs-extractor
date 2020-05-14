//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// This was largely copied from istio/istio/pkg/test/kube/kubectl.go

package kube

import (
	"fmt"
	"github.com/astralkn/k8s-logs-extractor/pkg/shell"
	"sync"
)

type kubectl struct {
	kubeConfig   string
	baseDir      string
	workDir      string
	workDirMutex sync.Mutex
}

// logs calls the logs command for the specified pod, with -c, if container is specified.
func (c *kubectl) logs(namespace string, pod string, container string, previousLog bool) (string, error) {
	cmd := fmt.Sprintf("kubectl logs %s %s %s %s %s",
		c.configArg(), namespaceArg(namespace), pod, containerArg(container), previousLogArg(previousLog))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) dumpInfo(outputDir, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl %s cluster-info dump %s --output=yaml --output-directory=%s",
		c.configArg(), namespaceArg(ns), outputDir)

	s, err := shell.Execute(true, cmd)

	if err != nil {
		return "", fmt.Errorf("%v: %s", err, s)
	}
	return s, nil
}

func (c *kubectl) crd(crd string) (string, error) {
	cmd := fmt.Sprintf("kubectl get crd %s %s",
		c.configArg(), crd)

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) configArg() string {
	return configArg(c.kubeConfig)
}

func (c *kubectl) describeCm(cm, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl describe cm %s %s %s",
		c.configArg(), cm, namespaceArg(ns))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) describeSVC(svc, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl describe svc %s %s %s",
		c.configArg(), svc, namespaceArg(ns))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) describeCRD(crd, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl describe crd %s %s %s",
		c.configArg(), crd, namespaceArg(ns))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) describeCR(cr, crd, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl describe %s %s %s %s",
		crd, cr, c.configArg(), namespaceArg(ns))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func (c *kubectl) pods(pod string, ns string) (string, error) {
	cmd := fmt.Sprintf("kubectl get pods %s %s %s",
		pod, c.configArg(), namespaceArg(ns))

	s, err := shell.Execute(true, cmd)

	if err == nil {
		return s, nil
	}

	return "", fmt.Errorf("%v: %s", err, s)
}

func configArg(kubeConfig string) string {
	if kubeConfig != "" {
		return fmt.Sprintf("--kubeconfig=%s", kubeConfig)
	}
	return ""
}

func namespaceArg(namespace string) string {
	if namespace == "all" {
		return "--all-namespaces"
	}
	if namespace != "" {
		return fmt.Sprintf("-n %s", namespace)
	}
	return ""
}

func containerArg(container string) string {
	if container != "" {
		return fmt.Sprintf("-c %s", container)
	}
	return ""
}

func previousLogArg(previous bool) string {
	if previous {
		return "-p"
	}
	return ""
}
