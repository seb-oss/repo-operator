# Using Repository resources

Before you can create repository resources, you need to [install repo-operator](installing.md)

* Edit [repository_v1beta1_repository.yaml](../config/samples/repository_v1beta1_repository.yaml) and change fields to your required values:
```
apiVersion: repository.storage.sebshift.io/v1beta1
kind: Repository
metadata:
  name: repo-test
spec:
  repotype: 'maven/docker/nuget/npm'
  users:
    - "user1"
    - "user2"
```
* **_name_** : Can be anything (but it is good to have names related to your project). If there is any existing repo with the name     service won't do anything.
* **_repotype_** : Choose one of the supported repotype  - 'maven'/'docker'/'nuget'/'npm' (Make sure you have remote repositories pre-configured for these repo types)
    * **_maven_** :  It will create snapshot and release repositories and add all maven remote repository to virtual repository.
    * **docker_** :  It will create docker repository and also create secret bind to your builder and default service account in namespace so that you can push your images directly into the Artifactory docker repository from kubernetes Image Build.
    * **_nuget/npm_** : It will create repositories and add all available remote repository of type to the virtual repository. 
    * **_Others_**: Not tested /supported as of now.
* **_users_**: specify all the users you want to give access to your repository (NOTE : User names should  be in small case). If Later you want to add/remove user you can make changes to the Repository object ("Resources → other resources → Choose Repository → your object → Edit Yaml ) and your permission object will be updated accordingly.
* Once the object is create successfully you can check the status of it by going to "Resources → other resources → Choose Repository → Edit Yaml → check statuscode it should be 200". you also get the repourl which you can  point to the repository.
* Never edit the repotype field after the object is created otherwise "Bad things will happen" :smiling_imp:
* If you delete the repository object, Operator will delete the repository and all the associated objects so please be very sure.

* Create instance of _**Repository**_  type 
```
kubectl apply -f getting_started/repository.yaml -n <namespace>
```
## Naming convention

Though William Shakespeare has rightly said – _"What's in a name?"_ but we still think that there should be a naming convention to follow.  
repo-operator imposes some naming convention while creating repositories and it's objects.
you should expect below names of objects created for a particular repo type  
* **_maven_**
    * *repositories*:     
        - '{name}-maven-snapshot-local'   
        - '{name}-maven-snapshot'   
        - '{name}-maven-release-local'   
        - '{name}-maven-release'   
    * *Permission*:  
        - '{name}-maven-repo-permission'   
* **_docker_**
    * *repositories*:   
         - '{name}-docker-local'
         - '{name}-docker'  
    * *Permission*:  
        - '{name}-maven-repo-permission'
    * *Repository Internal User*:  
        - '{name}-repo-user'
    * *Secret*
        - '{name}-repo-docker-secret' 
* **_Other repo type_** 
    * *repositories*:   
         - '{name}-{repotype}-local'
         - '{name}-{repotype}'  
    * *Permission*:  
        - '{name}-{repotype}-repo-permission'
        
:exclamation: Currently the naming is hardcoded into the operator but if you want something else, change the variables in code and build/deploy. 

:point_left: Back to [Home](../README.md#development)              