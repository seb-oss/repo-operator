package controllers

import (
	"context"
	repositoryv1beta1 "github.com/sebgroup/repo-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func Test_DockerRepositoryController(t *testing.T) {

	repository := &repositoryv1beta1.Repository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "repository.storage.sebshift.io/v1beta1",
			Kind:       "Repository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
		Spec: repositoryv1beta1.RepositorySpec{
			Repotype: "docker",
			Users:    []string{"testuser"},
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	saDefault := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	saBuilder := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "builder",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		repository,
		ns,
		saDefault,
		saBuilder,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(repositoryv1beta1.GroupVersion, repository)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Invoke the mock client
	a := mockRepositoryClient{}

	// Create a ReconcileMonitor object with the scheme, fake client and fake repo.
	r := &RepositoryReconciler{Client: cl, Log: ctrl.Log.WithName("test"), Scheme: s, rtc: &a}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
	}

	// Lets reconcile and setup the finalizer
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}
	var found bool
	f := repository.GetFinalizers()
	for _, fin := range f {
		if found = (fin == finalizer); !found {
			t.Errorf("reconcile did not set expected finalizer (%s), finalizers: %v", finalizer, f)
		}
	}

	// Reconcile again and go past finalizer create objects.
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}

	if repository.Status.Statuscode != 200 {
		t.Errorf("repository status does not have the expected values")
	}

	if repository.Status.Repourl != repositoryURL+"/test-repository-docker" {
		t.Errorf("repository  does not have the correct url name")
	}

}

func Test_MavenRepositoryController(t *testing.T) {

	repository := &repositoryv1beta1.Repository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "repository.storage.sebshift.io/v1beta1",
			Kind:       "Repository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
		Spec: repositoryv1beta1.RepositorySpec{
			Repotype: "maven",
			Users:    []string{"testuser"},
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	saDefault := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	saBuilder := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "builder",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		repository,
		ns,
		saDefault,
		saBuilder,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(repositoryv1beta1.GroupVersion, repository)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Invoke the mock client
	a := mockRepositoryClient{}

	// Create a ReconcileMonitor object with the scheme, fake client and fake repo.
	r := &RepositoryReconciler{Client: cl, Log: ctrl.Log.WithName("test"), Scheme: s, rtc: &a}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
	}

	// Lets reconcile and setup the finalizer
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}
	var found bool
	f := repository.GetFinalizers()
	for _, fin := range f {
		if found = (fin == finalizer); !found {
			t.Errorf("reconcile did not set expected finalizer (%s), finalizers: %v", finalizer, f)
		}
	}

	// Reconcile again and go past finalizer create objects.
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}

	if repository.Status.Statuscode != 200 {
		t.Errorf("repository status does not have the expected values")
	}

	if repository.Status.Repourl != repositoryURL+"/test-repository-maven"+releaseSuffix {
		t.Errorf("repository  does not have the correct url name")
	}

}

func Test_DefaultRepositoryController(t *testing.T) {

	repository := &repositoryv1beta1.Repository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "repository.storage.sebshift.io/v1beta1",
			Kind:       "Repository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
		Spec: repositoryv1beta1.RepositorySpec{
			Repotype: "npm",
			Users:    []string{"testuser"},
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	saDefault := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	saBuilder := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "builder",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		repository,
		ns,
		saDefault,
		saBuilder,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(repositoryv1beta1.GroupVersion, repository)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Invoke the mock client
	a := mockRepositoryClient{}

	// Create a ReconcileMonitor object with the scheme, fake client and fake repo.
	r := &RepositoryReconciler{Client: cl, Log: ctrl.Log.WithName("test"), Scheme: s, rtc: &a}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
	}

	// Lets reconcile and setup the finalizer
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}
	var found bool
	f := repository.GetFinalizers()
	for _, fin := range f {
		if found = (fin == finalizer); !found {
			t.Errorf("reconcile did not set expected finalizer (%s), finalizers: %v", finalizer, f)
		}
	}

	// Reconcile again and go past finalizer create objects.
	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}

	if repository.Status.Statuscode != 200 {
		t.Errorf("repository status does not have the expected values")
	}

	if repository.Status.Repourl != repositoryURL+"/test-repository-npm" {
		t.Errorf("repository  does not have the correct url name")
	}

}

func Test_OtherCleanupRepositoryController(t *testing.T) {
	deletionTimestamp := metav1.NewTime(time.Now())
	repository := &repositoryv1beta1.Repository{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "repository.storage.sebshift.io/v1beta1",
			Kind:       "Repository",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-repository",
			Namespace:         "test-namespace",
			DeletionTimestamp: &deletionTimestamp,
			Finalizers: []string{
				finalizer,
			},
		},
		Spec: repositoryv1beta1.RepositorySpec{
			Repotype: "npm",
			Users:    []string{"testuser"},
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	saDefault := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "default",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	saBuilder := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "builder",
			UID:             "12345",
			Namespace:       "test-namespace",
			ResourceVersion: "1",
		},
		Secrets:          nil,
		ImagePullSecrets: nil,
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		repository,
		ns,
		saDefault,
		saBuilder,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(repositoryv1beta1.GroupVersion, repository)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Invoke the mock client
	a := mockRepositoryClient{}

	// Create a ReconcileMonitor object with the scheme, fake client and fake repo.
	r := &RepositoryReconciler{Client: cl, Log: ctrl.Log.WithName("test"), Scheme: s, rtc: &a}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-repository",
			Namespace: "test-namespace",
		},
	}

	// Lets reconcile and setup the finalizer
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Get updated repository object
	err = cl.Get(context.TODO(), req.NamespacedName, repository)
	if err != nil {
		t.Fatalf("get repository: (%v)", err)
	}

	if repository.Status.Statuscode == 200 {
		t.Errorf("repository not deleted properly")
	}

}

type mockRepositoryClient struct{}

func (m *mockRepositoryClient) CreateRepositories(repoName string, repoType string, namespace string, statusCode int) (int, string, error) {
	return 200, "ok", nil
}

func (m *mockRepositoryClient) CreatePermission(reqName string, repoType string, namespace string, users []string, repositories []string) error {
	return nil
}

func (m *mockRepositoryClient) CreateRepositoryUser(reqName string) (string, int, string, error) {
	return "password", 200, "ok", nil
}

func (m *mockRepositoryClient) CleanupRepository(reqName string, repoType string, namespace string) error {
	return nil
}
