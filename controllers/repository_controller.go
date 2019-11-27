/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	repositoryv1beta1 "github.com/sebgroup/repo-operator/api/v1beta1"
	"github.com/sebgroup/repo-operator/pkg/repository"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_repository")
var repositoryURL string

// DockerConfig : config entry
type DockerConfig map[string]DockerConfigEntry

// DockerConfigJSON : auth struct
type DockerConfigJSON struct {
	Auths DockerConfig `json:"auths"`
}

//  interface
type rtInterface interface {
	CreateRepositories(repoName string, repoType string, namespace string, statusCode int) (int, string, error)
	CreatePermission(reqName string, repoType string, namespace string, users []string, repositories []string) error
	CreateRepositoryUser(reqName string) (string, int, string, error)
	CleanupRepository(reqName string, repoType string, namespace string) error
}

// DockerConfigEntry : dockerconfig struct structure
type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

const (
	ins                       = "instance.Namespace"
	rname                     = "req.Name"
	dockerRepoType            = "docker"
	mavenRepoType             = "maven"
	finalizer                 = "finalizer.repositories.sebshift.io"
	conflictState             = "Conflict"
	suffixPackageClassLocal   = "-local"
	suffixArtifactoryRepoUser = "-repo-user"
	suffixSecretName          = "-repo-docker-secret"
	snapshotSuffix            = "-snapshot"
	releaseSuffix             = "-release"
	failToInsertStatusCode    = "failed to insert status code"
)

// RepositoryReconciler reconciles a Repository object
type RepositoryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	rtc    rtInterface
}

func init() {
	repositoryURL = os.Getenv("REPOSITORY_URL")
}

// +kubebuilder:rbac:groups=repository.storage.sebshift.io,resources=repositories,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=repository.storage.sebshift.io,resources=repositories/status,verbs=get;update;patch

//Reconcile : Main reconcile function
func (r *RepositoryReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("repository", req.NamespacedName)

	reqLogger := r.Log.WithValues(ins, req.Namespace, rname, req.Name)
	reqLogger.Info("Reconciling Artifactory Repository")

	// Fetch the Repository instance
	instance := &repositoryv1beta1.Repository{}

	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	result, err, done := r.checkForDeletion(instance, reqLogger, req)
	if done {
		return result, err
	}

	switch instance.Spec.Repotype {
	case mavenRepoType:
		err := r.createMavenRepositoryObjects(err, req, instance, reqLogger)
		if err != nil {
			return ctrl.Result{}, err
		}
	case dockerRepoType:
		err := r.createDockerRepositoryObjects(req, instance, reqLogger)
		if err != nil {
			return ctrl.Result{}, err
		}

	default:
		err := r.createOtherRepositoryObjects(req, instance, reqLogger)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// All objects created successfully - don't requeue
	return ctrl.Result{}, nil
}

// Create Objects for Maven repository type
func (r *RepositoryReconciler) createMavenRepositoryObjects(err error, req ctrl.Request, instance *repositoryv1beta1.Repository, reqLogger logr.Logger) error {
	// Input received
	namespace := req.Namespace
	repositoryType := instance.Spec.Repotype
	statusCode := instance.Status.Statuscode

	// Naming standard defined below - change these values to meet your demands
	mavenSnapshotRepositoryName := req.Name + "-" + repositoryType + snapshotSuffix
	mavenReleaseRepositoryName := req.Name + "-" + repositoryType + releaseSuffix

	//Create maven snapshot Local & Virtual Artifactory repository
	code, status, err := r.rtc.CreateRepositories(mavenSnapshotRepositoryName, repositoryType, namespace, statusCode)
	if err != nil {
		return err
	}
	//Create maven release Local & Virtual Artifactory repository
	code, status, err = r.rtc.CreateRepositories(mavenReleaseRepositoryName, repositoryType, namespace, statusCode)
	if err != nil {
		return err
	}
	//Set status
	if code != instance.Status.Statuscode {
		instance.Status.Statuscode = code
		instance.Status.State = status
		instance.Status.Repourl = repositoryURL + "/" + req.Name + "-" + repositoryType + releaseSuffix
		err = r.setStatus(instance)
		if err != nil {
			reqLogger.Error(err, failToInsertStatusCode)
			return err
		}
	}

	// Create Permission Object
	if instance.Status.State != conflictState {
		repositories := []string{req.Name + "-" + repositoryType + snapshotSuffix + suffixPackageClassLocal, req.Name + "-" + repositoryType + releaseSuffix + suffixPackageClassLocal}
		err = r.rtc.CreatePermission(req.Name, repositoryType, namespace, instance.Spec.Users, repositories)
		// We failed to create the permission, requeue to try again
		if err != nil {
			return err
		}
	} else {
		reqLogger.Info("This instance is in conflict state - do not create permission object")
	}
	return nil
}

// Create Objects for Docker repository type
func (r *RepositoryReconciler) createDockerRepositoryObjects(req ctrl.Request, instance *repositoryv1beta1.Repository, reqLogger logr.Logger) error {
	// Input received
	namespace := req.Namespace
	repositoryType := instance.Spec.Repotype
	statusCode := instance.Status.Statuscode

	// Naming standard defined below - change these values to meet your demands
	dockerRepositoryName := req.Name + "-" + repositoryType

	//Create Local & Virtual Repository repository
	code, status, err := r.rtc.CreateRepositories(dockerRepositoryName, repositoryType, namespace, statusCode)
	if err != nil {
		return err
	}
	// Create required Repository docker objects
	err = r.createWiring(instance, req)
	// We failed to get the secret, requeue to try again
	if err != nil {
		return err
	}
	//Set status
	if code != instance.Status.Statuscode {
		instance.Status.Statuscode = code
		instance.Status.State = status
		instance.Status.Repourl = repositoryURL + "/" + req.Name + "-" + repositoryType
		err = r.setStatus(instance)
		if err != nil {
			reqLogger.Error(err, failToInsertStatusCode)
			return err
		}
	}
	// Create Permission Object
	if instance.Status.State != conflictState {
		repositories := []string{req.Name + "-" + repositoryType + suffixPackageClassLocal}
		users := append(instance.Spec.Users, req.Name+suffixArtifactoryRepoUser)
		err = r.rtc.CreatePermission(req.Name, repositoryType, namespace, users, repositories)
		// We failed to create the permission, requeue to try again
		if err != nil {
			return err
		}
	} else {
		reqLogger.Info("This instance is in conflict state - do not create permission object")
	}
	return nil
}

// Create Objects fro all the other type of the repos.
func (r *RepositoryReconciler) createOtherRepositoryObjects(req ctrl.Request, instance *repositoryv1beta1.Repository, reqLogger logr.Logger) error {
	// Input received
	repositoryType := instance.Spec.Repotype
	namespace := req.Namespace
	statusCode := instance.Status.Statuscode

	// Naming standard defined below - change these values to meet your demands
	otherRepositoryName := req.Name + "-" + repositoryType

	//Create Local & Virtual Artifactory repository
	code, status, err := r.rtc.CreateRepositories(otherRepositoryName, repositoryType, namespace, statusCode)
	if err != nil {
		return err
	}
	//Set status
	if code != instance.Status.Statuscode {
		instance.Status.Statuscode = code
		instance.Status.State = status
		instance.Status.Repourl = repositoryURL + "/" + req.Name + "-" + repositoryType
		err = r.setStatus(instance)
		if err != nil {
			reqLogger.Error(err, failToInsertStatusCode)
			return err
		}
	}
	// Create Permission Object
	if instance.Status.State != conflictState {
		repositories := []string{req.Name + "-" + repositoryType + suffixPackageClassLocal}
		err = r.rtc.CreatePermission(req.Name, repositoryType, namespace, instance.Spec.Users, repositories)
		// We failed to create the permission, requeue to try again
		if err != nil {
			return err
		}
	} else {
		reqLogger.Info("This instance is in conflict state - do not create permission object")
	}
	return nil
}

// It creates service account, secret and user permission objects
func (r *RepositoryReconciler) createWiring(instance *repositoryv1beta1.Repository, req ctrl.Request) error {
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	reqLogger := log.WithValues(ins, instance.Namespace, rname, req.Name)

	//User list
	users := instance.Spec.Users
	// Is this a request for docker repository type.
	// Create secret and permission
	reqLogger.Info("Creating user secret with permissions to be able to push to docker registry", "Namespace", instance.Namespace, "Name", instance.Name)
	secretFound := &corev1.Secret{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: req.Name + suffixSecretName, Namespace: instance.Namespace}, secretFound)
	if err != nil && errors.IsNotFound(err) {
		// Create artifactory internal User
		reqLogger.Info("Create artifactory internal User", "Namespace", instance.Namespace, "Name", instance.Name)
		rp, _, _, err := r.rtc.CreateRepositoryUser(req.Name)
		if err != nil {
			reqLogger.Error(err, "failed to create user")
			return err
		}
		// Add user to the list
		users = append(users, req.Name+suffixArtifactoryRepoUser)

		// Create Secret with user and password
		reqLogger.Info("Create Secret with user and password", "Namespace", instance.Namespace, "Name", instance.Name)
		reqLogger.Info("Creating a new secret", "Namespace", instance.Namespace, "Name", instance.Name)
		s := r.generateRepoSecret(instance.Namespace, req.Name+suffixSecretName, instance.Name+suffixArtifactoryRepoUser, rp)
		//Set Owner reference
		reqLogger.Info("Setting owner reference", "Namespace", instance.Namespace, "Name", instance.Name)
		err = setOwnerReference(instance, s, r.Scheme)
		if err != nil {
			reqLogger.Error(err, "Unable to set owner reference")
			return err
		}
		err = r.Create(context.TODO(), s)
		if err != nil {
			reqLogger.Error(err, "Creation of secret failed!")
			return err
		}

		// Add secret as pull secret to default service account
		reqLogger.Info("Add secret as pull secret to default service account and secret to builder account", "Namespace", instance.Namespace, "Name", instance.Name)
		// Get Default service account
		reqLogger.Info("Get Default service account", "Namespace", instance.Namespace, "Name", instance.Name)
		defaultSA := &corev1.ServiceAccount{}
		err = r.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: "default"}, defaultSA)
		if err != nil {
			return err
		}

		//Get builder service account
		reqLogger.Info("Get Builder service account", "Namespace", instance.Namespace, "Name", instance.Name)
		builder := &corev1.ServiceAccount{}
		err = r.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: "builder"}, builder)
		if err != nil {
			return err
		}

		reqLogger.Info("Link secret to default service account as pull secret", "Namespace", instance.Namespace, "Name", instance.Name)
		// Link secret to default service account as pull secret
		linked := linkImagePullSecret(defaultSA, req.Name+suffixSecretName)
		if linked {
			err = r.Update(context.TODO(), defaultSA)
			if err != nil {
				return err
			}
		}

		reqLogger.Info("link secret to builder service account", "Namespace", instance.Namespace, "Name", instance.Name)
		// Link secret to builder service account
		linked = linkSecret(builder, req.Name+suffixSecretName) || linked
		if linked {
			err = r.Update(context.TODO(), builder)
			if err != nil {
				return err
			}
		}
	} else if err != nil {
		reqLogger.Info("Error is :"+err.Error(), "found.Namespace", secretFound.Namespace, "found.Name", secretFound.Name)
		return err
	}
	return nil
}

// Check if the object is deleted; if yes then call the cleanup
func (r *RepositoryReconciler) checkForDeletion(instance *repositoryv1beta1.Repository, reqLogger logr.Logger, req ctrl.Request) (ctrl.Result, error, bool) {
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The instance is not being deleted, add finalizer if not already there
		reqLogger.Info("The instance is not being deleted, add finalizer if not already there")
		if !containsString(instance.ObjectMeta.Finalizers, finalizer) {
			result, e, b, done := r.addFinalizer(instance)
			if done {
				return result, e, b
			}
			return ctrl.Result{}, nil, true
		}
	} else {
		// The instance is being deleted
		reqLogger.Info("The instance is being deleted")
		if containsString(instance.ObjectMeta.Finalizers, finalizer) {
			// remove our finalizer from the list and update it.
			result, e, b, done := r.removeFinalizer(instance)
			if done {
				return result, e, b
			}
			// our finalizer is there, cleanup repositories in Artifactory
			if err := r.cleanupEverything(instance, req); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err, true
			}

		}
		// Return for garbage collection to do its job
		return ctrl.Result{}, nil, true
	}
	return ctrl.Result{}, nil, false
}

// Add finalizer if not already exist
func (r *RepositoryReconciler) addFinalizer(instance *repositoryv1beta1.Repository) (ctrl.Result, error, bool, bool) {
	instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizer)
	if err := r.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err, true, true
	}
	return ctrl.Result{}, nil, false, false
}

// Remove finalizer when object is deleted
func (r *RepositoryReconciler) removeFinalizer(instance *repositoryv1beta1.Repository) (ctrl.Result, error, bool, bool) {
	instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, finalizer)
	if err := r.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err, true, true
	}
	return ctrl.Result{}, nil, false, false
}

// It clean-up everything related to repositories
func (r *RepositoryReconciler) cleanupEverything(instance *repositoryv1beta1.Repository, req ctrl.Request) error {
	reqLogger := log.WithValues(ins, instance.Namespace, rname, req.Name)
	reqLogger.Info("cleanup Everything is called....")
	err := r.Get(context.TODO(), req.NamespacedName, instance)

	if instance.Status.State != conflictState {
		err = r.rtc.CleanupRepository(req.Name, instance.Spec.Repotype, req.Namespace)
		if err != nil {
			reqLogger.Error(err, "Cleanup failed!")
			//return err
		}
		if instance.Spec.Repotype == dockerRepoType {
			err := r.cleanUpWiring(err, instance, req)
			if err != nil {
				return err
			}
		}
	} else {
		reqLogger.Info("This instance is in conflict state - better not do anything just delete")
	}
	return nil
}

// Cleanup the the wiring done for docker repo type
func (r *RepositoryReconciler) cleanUpWiring(err error, instance *repositoryv1beta1.Repository, req ctrl.Request) error {
	// Get Default service account
	defaultSA := &corev1.ServiceAccount{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: "default"}, defaultSA)
	if err != nil {
		return err
	}
	//Get builder service account
	builderSA := &corev1.ServiceAccount{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: "builder"}, builderSA)
	if err != nil {
		return err
	}
	// Un-Link builder service secret
	builderSA = unLinkBuilderSASecret(builderSA, req.Name)
	err = r.Update(context.TODO(), builderSA)
	if err != nil {
		return err
	}
	// Un-Link Image pull secret
	defaultSA = unLinkDefaultSAPullSecret(defaultSA, req.Name)
	err = r.Update(context.TODO(), defaultSA)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager : setup manager
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.rtc = repository.NewRepositoryClient()
	return ctrl.NewControllerManagedBy(mgr).
		For(&repositoryv1beta1.Repository{}).
		Complete(r)
}

// Set status in the repository object status field
func (r *RepositoryReconciler) setStatus(instance *repositoryv1beta1.Repository) error {
	// k8s 1.9 treats status as just another part of the object:
	return r.Update(context.TODO(), instance)
}
