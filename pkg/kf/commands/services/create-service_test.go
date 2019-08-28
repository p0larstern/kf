// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	servicecatalogclientfake "github.com/google/kf/pkg/client/servicecatalog/clientset/versioned/fake"
	"github.com/google/kf/pkg/kf/commands/config"
	servicescmd "github.com/google/kf/pkg/kf/commands/services"
	"github.com/google/kf/pkg/kf/commands/utils"
	"github.com/google/kf/pkg/kf/testutil"
	servicecatalogv1beta1 "github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
)

func TestNewCreateServiceCommand(t *testing.T) {

	plan := servicecatalogv1beta1.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-service-free",
			Namespace: "custom-ns",
		},
		Spec: servicecatalogv1beta1.ServicePlanSpec{
			CommonServicePlanSpec: servicecatalogv1beta1.CommonServicePlanSpec{
				ExternalName: "free",
			},
			ServiceClassRef: servicecatalogv1beta1.LocalObjectReference{
				Name: "db-service",
			},
		},
	}

	planList := &servicecatalogv1beta1.ServicePlanList{
		Items: []servicecatalogv1beta1.ServicePlan{
			plan,
		},
	}

	cases := map[string]struct {
		Args      []string
		Setup     func(t *testing.T) *servicecatalogclientfake.Clientset
		Namespace string

		ExpectedErr     error
		ExpectedStrings []string
	}{
		"too few params": {
			Args:        []string{},
			ExpectedErr: errors.New("accepts 3 arg(s), received 0"),
		},
		"command params get passed correctly": {
			Args:      []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				return servicecatalogclientfake.NewSimpleClientset(planList)
			},
			ExpectedStrings: []string{
				"Name:    mydb",
				"ram_gb",
			},
		},
		"empty namespace": {
			Args:        []string{"db-service", "free", "mydb", `--config={"ram_gb":4}`},
			ExpectedErr: errors.New(utils.EmptyNamespaceError),
		},
		"defaults config": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				return servicecatalogclientfake.NewSimpleClientset(dummyServerInstance("mydb"))
			},
		},
		"bad path": {
			Args:        []string{"db-service", "free", "mydb", `--config=/some/bad/path`},
			Namespace:   "custom-ns",
			ExpectedErr: errors.New("couldn't read file: open /some/bad/path: no such file or directory"),
		},
		"bad server call": {
			Args:      []string{"db-service", "free", "mydb"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T) *servicecatalogclientfake.Clientset {
				client := servicecatalogclientfake.NewSimpleClientset(dummyServerInstance("mydb"))
				client.AddReactor("*", "*", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("server-call-error")
				})
				return client
			},
			ExpectedErr: errors.New("server-call-error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {

			buf := new(bytes.Buffer)
			p := &config.KfParams{
				Namespace: tc.Namespace,
			}

			var client *servicecatalogclientfake.Clientset
			if tc.Setup != nil {
				client = tc.Setup(t)
			} else {
				client = servicecatalogclientfake.NewSimpleClientset()
			}

			cmd := servicescmd.NewCreateServiceCommand(p, client)
			fmt.Fprintf(os.Stderr, "%s", buf.String())
			cmd.SetOutput(buf)
			cmd.SetArgs(tc.Args)
			_, actualErr := cmd.ExecuteC()
			if tc.ExpectedErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectedErr, actualErr)
			}

			testutil.AssertContainsAll(t, buf.String(), tc.ExpectedStrings)
		})
	}
}

/*
func TestClient_CreateService(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		InstanceName string
		ServiceName  string
		PlanName     string
		Options      []CreateServiceOption
		ProvisionErr error

		ExpectErr error
	}{
		"default values": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			Options:      []CreateServiceOption{},
			ExpectErr:    nil,
		},
		"custom values": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			Options: []CreateServiceOption{
				WithCreateServiceNamespace("custom-namespace"),
				WithCreateServiceParams(map[string]interface{}{"foo": 33}),
			},
			ExpectErr: nil,
		},
		"error in provision": {
			InstanceName: "instance-name",
			ServiceName:  "service-name",
			PlanName:     "plan-name",
			ProvisionErr: errors.New("provision-err"),
			ExpectErr:    errors.New("provision-err"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			fakeClient := &servicecatalogfakes.FakeSvcatClient{}

			fakeClient.ProvisionStub = func(instanceName, className, planName string, opts *servicecatalog.ProvisionOptions) (*v1beta1.ServiceInstance, error) {
				expectedCfg := CreateServiceOptionDefaults().Extend(tc.Options).toConfig()

				testutil.AssertEqual(t, "instanceName", tc.InstanceName, instanceName)
				testutil.AssertEqual(t, "className", tc.ServiceName, className)
				testutil.AssertEqual(t, "planName", tc.PlanName, planName)
				testutil.AssertEqual(t, "opts.namespace", expectedCfg.Namespace, opts.Namespace)
				testutil.AssertEqual(t, "opts.params", expectedCfg.Params, opts.Params)

				return nil, tc.ProvisionErr
			}

			client := NewClient(func(ns string) servicecatalog.SvcatClient {
				return fakeClient
			})

			_, actualErr := client.CreateService(tc.InstanceName, tc.ServiceName, tc.PlanName, tc.Options...)
			if tc.ExpectErr != nil || actualErr != nil {
				testutil.AssertErrorsEqual(t, tc.ExpectErr, actualErr)

				return
			}

			testutil.AssertEqual(t, "calls to provision", 1, fakeClient.ProvisionCallCount())
		})
	}
}
*/
