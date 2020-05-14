//  Copyright 2018 Istio Authors
//
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

package kube

import (
	"fmt"
	kubeApiCore "k8s.io/api/core/v1"
	kubeExtClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeApiMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	kubeClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Needed for auth
	"k8s.io/client-go/rest"
)

// Accessor is a helper for accessing Kubernetes programmatically. It bundles some of the high-level
// operations that is frequently used by the test framework.
type Accessor struct {
	restConfig *rest.Config
	ctl        *kubectl
	set        *kubeClient.Clientset
	extSet     *kubeExtClient.Clientset
	dynClient  dynamic.Interface
}

// NewAccessor returns a new instance of an accessor.
func NewAccessor(kubeConfig string, baseWorkDir string) (*Accessor, error) {
	restConfig, err := BuildClientConfig(kubeConfig, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config. %v", err)
	}
	restConfig.APIPath = "/api"
	restConfig.GroupVersion = &kubeApiCore.SchemeGroupVersion
	restConfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}

	set, err := kubeClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	extSet, err := kubeExtClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &Accessor{
		restConfig: restConfig,
		ctl: &kubectl{
			kubeConfig: kubeConfig,
			baseDir:    baseWorkDir,
		},
		set:       set,
		extSet:    extSet,
		dynClient: dynClient,
	}, nil
}

func (a *Accessor) GetPods(pod, ns string) (string, error) {
	return a.ctl.pods(pod, ns)
}

// Logs calls the logs command for the specified pod, with -c, if container is specified.
func (a *Accessor) Logs(namespace string, pod string, container string, previousLog bool) (string, error) {
	return a.ctl.logs(namespace, pod, container, previousLog)
}

func (a *Accessor) DumpInfo(outputDir, ns string) (string, error) {
	return a.ctl.dumpInfo(outputDir, ns)
}

func (a *Accessor) GetNamespaces() ([]kubeApiCore.Namespace, error) {
	var opts kubeApiMeta.ListOptions
	n, err := a.set.CoreV1().Namespaces().List(opts)
	if err != nil {
		return nil, err
	}
	return n.Items, nil
}

func (a *Accessor) DescribePod(pod, ns string) (string, error) {
	return a.ctl.describeCm(pod, ns)
}

func (a *Accessor) DescribeCM(cm, ns string) (string, error) {
	return a.ctl.describeCm(cm, ns)
}

func (a *Accessor) DescribeSVC(svc, ns string) (string, error) {
	return a.ctl.describeSVC(svc, ns)
}

func (a *Accessor) DescribeCRD(crd, ns string) (string, error) {
	return a.ctl.describeCRD(crd, ns)
}

func (a *Accessor) DescribeCR(cr, crd, ns string) (string, error) {
	return a.ctl.describeCR(cr, crd, ns)
}
