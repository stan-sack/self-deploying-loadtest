// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kargo

import (
	"flag"
	"fmt"
	"io"
)

var (
	apiHost          string
	cpuLimit         string
	cpuRequest       string
	memoryLimit      string
	memoryRequest    string
	namespace        string
	EnableKubernetes bool
)

func init() {
	flag.StringVar(&apiHost, "api-host", "127.0.0.1:8001", "Kubernetes API server")
	flag.StringVar(&cpuLimit, "cpu-limit", "100m", "Max CPU in milicores")
	flag.StringVar(&cpuRequest, "cpu-request", "100m", "Min CPU in milicores")
	flag.StringVar(&memoryLimit, "memory-limit", "64M", "Max memory in MB")
	flag.StringVar(&memoryRequest, "memory-request", "64M", "Min memory in MB")
	flag.StringVar(&namespace, "namespace", "default", "The Kubernetes namespace.")
	flag.BoolVar(&EnableKubernetes, "kubernetes", false, "Deploy to Kubernetes.")
}

type DeploymentConfig struct {
	Annotations   map[string]string
	Args          []string
	Env           map[string]string
	BinaryURL     string
	cpuRequest    string
	cpuLimit      string
	memoryRequest string
	memoryLimit   string
	Name          string
	Replicas      int
	Namespace     string
	Labels        map[string]string
	Sidecars      []Container
	InitSidecars  []Container
	ConfigMaps    []ConfigMap
	Volumes       []Volume
	DaemonSets    []Container
}

type DeploymentManager struct {
	apiHost string
	config  DeploymentConfig
}

func New() *DeploymentManager {
	return &DeploymentManager{apiHost: apiHost}
}

func (dm *DeploymentManager) Create(config DeploymentConfig) error {
	config.cpuRequest = cpuRequest
	config.cpuLimit = cpuLimit
	config.memoryRequest = memoryRequest
	config.memoryLimit = memoryLimit
	config.Namespace = namespace

	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	dm.config = config

	fmt.Printf("Creating %s ReplicaSet...\n", config.Name)
	createConfigMaps(dm.config)
	createDaemonSets(dm.config)
	return createReplicaSet(dm.config)
}

func (dm *DeploymentManager) Scale(config DeploymentConfig, n int) error {
	return scaleReplicaSet(config.Namespace, config.Name, n)
}

func (dm *DeploymentManager) Delete() error {
	fmt.Printf("Deleting %s ReplicaSet...\n", dm.config.Name)
	deleteConfigMaps(dm.config)
	deleteDaemonSets(dm.config)
	return deleteReplicaSet(dm.config)
}

func (dm *DeploymentManager) Logs(w io.Writer) error {
	return getLogs(dm.config, w)
}
