package controllers

import (
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestRepositoryReconciler_generateRepoSecret(t *testing.T) {
	type fields struct {
		Client client.Client
		Log    logr.Logger
		Scheme *runtime.Scheme
		rtc    rtInterface
	}
	type args struct {
		n   string
		req string
		u   string
		p   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v1.Secret
	}{
		{
			name: "Test Secret creation",
			args: args{n: "test-namespace", req: "test-req-repo-docker-secret", u: "username", p: "password"},
			want: &v1.Secret{
				ObjectMeta: controllerruntime.ObjectMeta{
					Name:      "test-req-repo-docker-secret",
					Namespace: "test-namespace",
				},
				Type: "kubernetes.io/dockerconfigjson",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RepositoryReconciler{
				Client: tt.fields.Client,
				Log:    tt.fields.Log,
				Scheme: tt.fields.Scheme,
				rtc:    tt.fields.rtc,
			}
			if got := r.generateRepoSecret(tt.args.n, tt.args.req, tt.args.u, tt.args.p); !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("generateRepoSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linkImagePullSecret(t *testing.T) {
	type args struct {
		sa     *v1.ServiceAccount
		secret string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test link imagePull secret does not exist",
			args: args{
				sa: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "default",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets:          nil,
					ImagePullSecrets: nil,
				},
				secret: "pull-secret",
			},
			want: true,
		},
		{
			name: "Test link imagePull secret already exist",
			args: args{
				sa: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "default",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets: nil,
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "pull-secret",
						},
					},
				},
				secret: "pull-secret",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := linkImagePullSecret(tt.args.sa, tt.args.secret); got != tt.want {
				t.Errorf("linkImagePullSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_linkSecret(t *testing.T) {
	type args struct {
		sa     *v1.ServiceAccount
		secret string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test link secret does not exist",
			args: args{
				sa: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "default",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets:          nil,
					ImagePullSecrets: nil,
				},
				secret: "does-not-exist-secret",
			},
			want: true,
		},
		{
			name: "Test link secret already exist",
			args: args{
				sa: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "default",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets: []v1.ObjectReference{
						{
							Name: "exist-secret",
						},
					},
					ImagePullSecrets: nil,
				},
				secret: "exist-secret",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := linkSecret(tt.args.sa, tt.args.secret); got != tt.want {
				t.Errorf("linkSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unLinkBuilderSASecret(t *testing.T) {
	type args struct {
		builderSA *v1.ServiceAccount
		reqName   string
	}
	tests := []struct {
		name string
		args args
		want *v1.ServiceAccount
	}{
		{
			name: "UnLink the secret from builder SA",
			args: args{
				builderSA: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "builder",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets: []v1.ObjectReference{
						{
							Name: "test-repo-docker-secret",
						},
						{
							Name: "other-docker-secret",
						},
					},
					ImagePullSecrets: nil,
				},
				reqName: "test",
			},
			want: &v1.ServiceAccount{
				TypeMeta: v12.TypeMeta{},
				ObjectMeta: v12.ObjectMeta{
					Name:            "builder",
					UID:             "12345",
					Namespace:       "default-namespace",
					ResourceVersion: "1",
				},
				Secrets: []v1.ObjectReference{
					{
						Name: "other-docker-secret",
					},
				},
				ImagePullSecrets: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unLinkBuilderSASecret(tt.args.builderSA, tt.args.reqName)
			for _, s := range got.Secrets {
				if s.Name == "test-repo-docker-secret" {
					t.Errorf("unLinkBuilderSASecret() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_unLinkDefaultSAPullSecret(t *testing.T) {
	type args struct {
		defaultSA *v1.ServiceAccount
		reqName   string
	}
	tests := []struct {
		name string
		args args
		want *v1.ServiceAccount
	}{
		{
			name: "Unlink the pull secret from default SA",
			args: args{
				defaultSA: &v1.ServiceAccount{
					TypeMeta: v12.TypeMeta{},
					ObjectMeta: v12.ObjectMeta{
						Name:            "default",
						UID:             "12345",
						Namespace:       "default-namespace",
						ResourceVersion: "1",
					},
					Secrets: nil,
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "test-repo-docker-secret",
						},
						{
							Name: "pull-secret",
						},
					},
				},
				reqName: "test",
			},
			want: &v1.ServiceAccount{
				TypeMeta: v12.TypeMeta{},
				ObjectMeta: v12.ObjectMeta{
					Name:            "default",
					UID:             "12345",
					Namespace:       "default-namespace",
					ResourceVersion: "1",
				},
				Secrets: nil,
				ImagePullSecrets: []v1.LocalObjectReference{
					{
						Name: "pull-secret",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unLinkDefaultSAPullSecret(tt.args.defaultSA, tt.args.reqName)
			for _, s := range got.ImagePullSecrets {
				if s.Name == "test-repo-docker-secret" {
					t.Errorf("unLinkDefaultSAPullSecret() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_containsString(t *testing.T) {
	type args struct {
		slice []string
		s     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test contain string in a slice of strings",
			args: args{
				slice: []string{
					"string1",
					"string2",
				},
				s: "string2",
			},
			want: true,
		},
		{
			name: "Test does not contain string in a slice of strings",
			args: args{
				slice: []string{
					"string1",
					"string2",
				},
				s: "string3",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsString(tt.args.slice, tt.args.s); got != tt.want {
				t.Errorf("containsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeString(t *testing.T) {
	type args struct {
		slice []string
		s     string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
	}{
		{
			name: "Test remove string in a slice of strings",
			args: args{
				slice: []string{
					"string1",
					"string2",
				},
				s: "string2",
			},
			wantResult: []string{
				"string1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := removeString(tt.args.slice, tt.args.s); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("removeString() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
