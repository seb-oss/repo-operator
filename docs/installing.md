# Installing repo-operator

:white_check_mark: Pre-Requisite
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and access to a [Kubernetes](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) cluster version 1.11 or later.
* Go 1.13
* Docker
* Make
* Artifactory-Pro (> 6.12.2)


To be able to install repo-operator in your kubernetes environment, follow below steps:
1. _**Clone this repository to your local drive**_ 
2. _**Install CRD into the Kubernetes cluster**_ 
```
kubectl apply -f config/crd/bases/repository.storage.sebshift.io_repositories.yaml
```   
> ¤ You need to have cluster-admin role to perform this task.        
  ¤ This step will install CustomResourceDefinition **repositories.repository.storage.sebshift.io** of kind **Repository** in the cluster.   
  ¤ We need to create CRD before we install repo-operator in the cluster.   
  
3. _**Build Docker image**_
```
make docker-build
``` 
> ¤ Set **GOPROXY** variable in the dockerfile if you are running behind firewall.   
  ¤ Set proxy in docker for windows if running behind firewall.   
  ¤ After you build and tag the image you need to push it to docker registry. Registry should be accessible from your kubernetes cluster.   
4. _**Create a namespace or select existing namespace to run repo-operator.**_
```
kubectl create ns <namespace>
```
5. _**Create Service account**_
```
kubectl apply -f deployment/sa.yaml -n <namespace>
```
> ¤ This service account will be used by operator to watch CRD resource in the cluster.     
6. _**Create Role and Role Binding**_
```
Before running edit role_binding.yaml and put the name of your repo-operator namespace

kubectl apply -f deployment/role.yaml
kubectl apply -f deployment/role_binding.yaml -n <namespace>
```
> ¤ A new cluster role is created to watch the CRD resource. This limits the permissions of SA when we create the role binding.    
  ¤ Role binding is required to give service account ClusterRole we created so that it can watch the Repository resource in all the namespace of the cluster.   
7. _**Create non cluster roles**_
 ```
 kubectl apply -f deployment/role-non-cluster-admin-edit.yaml
 kubectl apply -f deployment/role-non-cluster-admin-view.yaml
 ```
> ¤ This will allow non cluster-admin to create instance of CRD into their namespaces.    
8. _**Create Artifactory Admin User**_   
Either create an internal user or use a user whose password doesn't expire.  
9. _**Create Artifactory secret**_  
 ```
 Enter username and password created in above step *base64 encoded* in file deployment/secret.yaml  
 kubectl apply -f deployment/secret.yaml -n <namespace>
 ```  
> ¤ This secret will be mounted on operator and used to connect with Artifactory using Rest api.  
10. _**Create Deployment**_
 ```
  Replace values of env variables in deployment/repo-operator.yaml  
    * REPOSITORY_URL
    * image
  kubectl apply -f deployment/repo-operator.yaml -n <namespace>
 ```
> ¤ This will start deployment and start repo-operator watching for Repository object in the cluster

:point_right: You're now all set up to [use repository resources](using.md)

:point_left: Back to [Home](../README.md)      
   
