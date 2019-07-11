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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

var (
	replicasetsEndpoint = "/apis/extensions/v1beta1/namespaces/%s/replicasets"
	replicasetEndpoint  = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s"
	scaleEndpoint       = "/apis/extensions/v1beta1/namespaces/%s/replicasets/%s/scale"
	daemonSetsEndpoint  = "/apis/extensions/v1beta1/namespaces/%s/daemonsets"
	daemonSetEndpoint   = "/apis/extensions/v1beta1/namespaces/%s/daemonsets/%s"
	logsEndpoint        = "/api/v1/namespaces/%s/pods/%s/log"
	podsEndpoint        = "/api/v1/namespaces/%s/pods"
	configMapsEndpoint  = "/api/v1/namespaces/%s/configmaps"
	configMapEndpoint   = "/api/v1/namespaces/%s/configmaps/%s"
)

var ErrNotExist = errors.New("does not exist")

func getPods(namespace, labelSelector string) (*PodList, error) {
	var podList *PodList

	v := url.Values{}
	v.Set("labelSelector", labelSelector)

	path := fmt.Sprintf(podsEndpoint, namespace)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:     apiHost,
			Path:     path,
			Scheme:   "http",
			RawQuery: v.Encode(),
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		fmt.Println("No pods found using selector: ", labelSelector)
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get pods error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&podList)
	if err != nil {
		return nil, err
	}
	return podList, nil

}

func getLogs(config DeploymentConfig, w io.Writer) error {
	time.Sleep(10 * time.Second)
	rs, err := getReplicaSet(config.Namespace, config.Name)
	if err != nil {
		return err
	}

	var labelSelector bytes.Buffer
	for key, value := range rs.Spec.Selector.MatchLabels {
		labelSelector.WriteString(fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := getPods(config.Namespace, labelSelector.String())
	if err != nil {
		return err
	}

	go func(podList *PodList) {
		for {
			time.Sleep(60 * time.Second)
			newPodList, err := getPods(config.Namespace, labelSelector.String())
			if err != nil {
				fmt.Println(err)
			} else {
				podList = newPodList
			}
		}
	}(podList)

	for _, pod := range podList.Items {
		v := url.Values{}
		v.Set("follow", "true")
		v.Set("container", config.Name)
		v.Set("container", config.Name)

		path := fmt.Sprintf(logsEndpoint, config.Namespace, pod.Metadata.Name)
		request := &http.Request{
			Header: make(http.Header),
			Method: http.MethodGet,
			URL: &url.URL{
				Host:     apiHost,
				Path:     path,
				Scheme:   "http",
				RawQuery: v.Encode(),
			},
		}
		request.Header.Set("Accept", "application/json, */*")

		go func() {
			for {
				resp, err := http.DefaultClient.Do(request)
				if err != nil {
					fmt.Println(err)
					time.Sleep(5 * time.Second)
					continue
				}

				if resp.StatusCode == 404 {
					data, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err)
						time.Sleep(5 * time.Second)
						continue
					}
					fmt.Println(string(data))
					fmt.Println("GET pod logs error: ", ErrNotExist)
					time.Sleep(5 * time.Second)
					continue
				}
				if resp.StatusCode != 200 {
					fmt.Println(errors.New("Get replica set error non 200 reponse: " + resp.Status))
					time.Sleep(5 * time.Second)
					continue
				}

				if _, err := io.Copy(w, resp.Body); err != nil {
					fmt.Println(err)
				}
			}
		}()
	}

	return nil
}

func getReplicaSet(namespace, name string) (*ReplicaSet, error) {
	var rs ReplicaSet

	path := fmt.Sprintf(replicasetEndpoint, namespace, name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get deployment error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func getScale(namespace, name string) (*Scale, error) {
	var scale Scale

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrNotExist
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Get scale error non 200 reponse: " + resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&scale)
	if err != nil {
		return nil, err
	}
	return &scale, nil
}

func scaleReplicaSet(namespace, name string, replicas int) error {
	scale, err := getScale(namespace, name)
	if err != nil {
		return err
	}
	scale.Spec.Replicas = int64(replicas)

	var b []byte
	body := bytes.NewBuffer(b)
	err = json.NewEncoder(body).Encode(scale)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(scaleEndpoint, namespace, name)
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPut,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return errors.New("Scale ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}

func deleteReplicaSet(config DeploymentConfig) error {
	err := scaleReplicaSet(config.Namespace, config.Name, 0)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetEndpoint, config.Namespace, config.Name)
	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodDelete,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Accept", "application/json, */*")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return ErrNotExist
	}
	if resp.StatusCode != 200 {
		return errors.New("Delete ReplicaSet error non 200 reponse: " + resp.Status)
	}

	return nil
}

func deleteConfigMaps(config DeploymentConfig) error {
	for _, cm := range config.ConfigMaps {
		path := fmt.Sprintf(configMapEndpoint, config.Namespace, cm.Metadata.Name)
		request := &http.Request{
			Header: make(http.Header),
			Method: http.MethodDelete,
			URL: &url.URL{
				Host:   apiHost,
				Path:   path,
				Scheme: "http",
			},
		}
		request.Header.Set("Accept", "application/json, */*")

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		if resp.StatusCode == 404 {
			return ErrNotExist
		}
		if resp.StatusCode != 200 {
			return errors.New("Delete ConfigMap error non 200 reponse: " + resp.Status)
		}
	}

	return nil
}

func deleteDaemonSets(config DeploymentConfig) error {
	for _, ds := range config.DaemonSets {
		path := fmt.Sprintf(daemonSetEndpoint, config.Namespace, ds.Name)
		request := &http.Request{
			Header: make(http.Header),
			Method: http.MethodDelete,
			URL: &url.URL{
				Host:   apiHost,
				Path:   path,
				Scheme: "http",
			},
		}
		request.Header.Set("Accept", "application/json, */*")

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		if resp.StatusCode == 404 {
			return ErrNotExist
		}
		if resp.StatusCode != 200 {
			return errors.New("Delete Damonset error non 200 reponse: " + resp.Status)
		}
	}

	return nil
}

func createConfigMaps(config DeploymentConfig) error {
	for _, cm := range config.ConfigMaps {
		var b []byte
		body := bytes.NewBuffer(b)
		err := json.NewEncoder(body).Encode(cm)
		if err != nil {
			return err
		}

		path := fmt.Sprintf(configMapsEndpoint, config.Namespace)
		request := &http.Request{
			Body:          ioutil.NopCloser(body),
			ContentLength: int64(body.Len()),
			Header:        make(http.Header),
			Method:        http.MethodPost,
			URL: &url.URL{
				Host:   apiHost,
				Path:   path,
				Scheme: "http",
			},
		}
		request.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		if resp.StatusCode != 201 {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return errors.New("ConfigMap: Unexpected HTTP status code" + resp.Status)
		}
	}

	return nil
}

func createDaemonSets(config DeploymentConfig) error {
	for _, ds := range config.DaemonSets {
		var b []byte
		body := bytes.NewBuffer(b)
		err := json.NewEncoder(body).Encode(ds)
		if err != nil {
			return err
		}

		path := fmt.Sprintf(daemonSetsEndpoint, config.Namespace)
		request := &http.Request{
			Body:          ioutil.NopCloser(body),
			ContentLength: int64(body.Len()),
			Header:        make(http.Header),
			Method:        http.MethodPost,
			URL: &url.URL{
				Host:   apiHost,
				Path:   path,
				Scheme: "http",
			},
		}
		request.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		if resp.StatusCode != 201 {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return errors.New("DaemonSet: Unexpected HTTP status code" + resp.Status)
		}
	}

	return nil
}

func createReplicaSet(config DeploymentConfig) error {

	volumes := append(config.Volumes, Volume{
		Name:         "bin",
		VolumeSource: VolumeSource{},
	})

	volumeMounts := make([]VolumeMount, 0)
	volumeMounts = append(volumeMounts, VolumeMount{
		Name:      "bin",
		MountPath: "/opt/bin",
	})

	container := Container{
		Args:            config.Args,
		Command:         []string{filepath.Join("/opt/bin", config.Name)},
		Image:           "alpine",
		ImagePullPolicy: "Always",
		Name:            config.Name,
		VolumeMounts:    volumeMounts,
	}

	resourceLimits := make(ResourceList)
	if config.cpuLimit != "" {
		resourceLimits["cpu"] = config.cpuLimit
	}
	if config.memoryLimit != "" {
		resourceLimits["memory"] = config.memoryLimit
	}

	resourceRequests := make(ResourceList)
	if config.cpuRequest != "" {
		resourceRequests["cpu"] = config.cpuRequest
	}
	if config.memoryRequest != "" {
		resourceRequests["memory"] = config.memoryRequest
	}

	if len(resourceLimits) > 0 {
		container.Resources.Limits = resourceLimits
	}
	if len(resourceRequests) > 0 {
		container.Resources.Requests = resourceRequests
	}

	if len(config.Env) > 0 {
		env := make([]EnvVar, 0)
		for name, value := range config.Env {
			env = append(env, EnvVar{Name: name, Value: value})
		}
		container.Env = env
	}

	annotations := config.Annotations

	binaryPath := filepath.Join("/opt/bin", config.Name)
	initContainer0 := Container{
		Name:            "install",
		Image:           "alpine",
		ImagePullPolicy: "Always",
		Command:         []string{"wget", "-O", binaryPath, config.BinaryURL},
		VolumeMounts: []VolumeMount{
			VolumeMount{
				Name:      "bin",
				MountPath: "/opt/bin",
			},
		},
	}

	initContainer1 := Container{
		Name:            "configure",
		Image:           "alpine",
		ImagePullPolicy: "Always",
		Command:         []string{"chmod", "+x", binaryPath},
		VolumeMounts: []VolumeMount{
			VolumeMount{
				Name:      "bin",
				MountPath: "/opt/bin",
			},
		},
	}

	initContainers := append(config.InitSidecars, initContainer0, initContainer1)

	config.Labels["run"] = config.Name

	rs := ReplicaSet{
		ApiVersion: "extensions/v1beta1",
		Kind:       "ReplicaSet",
		Metadata: Metadata{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: ReplicaSetSpec{
			Replicas: int64(config.Replicas),
			Selector: LabelSelector{
				MatchLabels: config.Labels,
			},
			Template: PodTemplate{
				Metadata: Metadata{
					Annotations: annotations,
					Labels:      config.Labels,
				},
				Spec: PodSpec{
					Containers:     append(config.Sidecars, container),
					InitContainers: initContainers,
					Volumes:        volumes,
				},
			},
		},
	}

	var b []byte
	body := bytes.NewBuffer(b)
	err := json.NewEncoder(body).Encode(rs)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(replicasetsEndpoint, config.Namespace)
	request := &http.Request{
		Body:          ioutil.NopCloser(body),
		ContentLength: int64(body.Len()),
		Header:        make(http.Header),
		Method:        http.MethodPost,
		URL: &url.URL{
			Host:   apiHost,
			Path:   path,
			Scheme: "http",
		},
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return errors.New("ReplicaSet: Unexpected HTTP status code" + resp.Status)
	}

	return nil
}
