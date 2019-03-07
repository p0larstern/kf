package kf

import (
	"errors"
	"fmt"
	"testing"

	cserving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeleteCommand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name              string
		namespace         string
		appName           string
		wantErr           error
		servingFactoryErr error
		serviceDeleteErr  error
	}{
		{
			name:      "deletes given app in namespace",
			namespace: "some-namespace",
			appName:   "some-app",
		},
		{
			name:    "deletes given app in the default namespace",
			appName: "some-app",
		},
		{
			name:    "empty app name, returns error",
			wantErr: errors.New("invalid app name"),
		},
		{
			name:              "serving factory error",
			wantErr:           errors.New("some error"),
			servingFactoryErr: errors.New("some error"),
			appName:           "some-app",
		},
		{
			name:             "service delete error",
			wantErr:          errors.New("some error"),
			serviceDeleteErr: errors.New("some error"),
			appName:          "some-app",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fake.FakeServingV1alpha1{
				Fake: &ktesting.Fake{},
			}

			expectedNamespace := tc.namespace
			if tc.namespace == "" {
				expectedNamespace = "default"
			}

			called := false
			fake.AddReactor("*", "*", ktesting.ReactionFunc(func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				called = true
				if action.GetNamespace() != expectedNamespace {
					t.Fatalf("wanted namespace: %s, got: %s", expectedNamespace, action.GetNamespace())
				}

				if !action.Matches("delete", "services") {
					t.Fatal("wrong action")
				}

				if gn := action.(ktesting.DeleteAction).GetName(); gn != tc.appName {
					t.Fatalf("wanted app name %s, got %s", tc.appName, gn)
				}

				return tc.serviceDeleteErr != nil, nil, tc.serviceDeleteErr
			}))

			d := NewDeleter(func() (cserving.ServingV1alpha1Interface, error) {
				return fake, tc.servingFactoryErr
			})

			var opts []DeleteOption
			if tc.namespace != "" {
				opts = append(opts, WithDeleteNamespace(tc.namespace))
			}

			gotErr := d.Delete(tc.appName, opts...)
			if tc.wantErr != nil || gotErr != nil {
				if fmt.Sprint(tc.wantErr) != fmt.Sprint(gotErr) {
					t.Fatalf("wanted err: %v, got: %v", tc.wantErr, gotErr)
				}

				return
			}

			if !called {
				t.Fatal("Reactor was not invoked")
			}
		})
	}
}
