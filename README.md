# **Dummy Operator** #
dummy-operator is a simple kubernetes custom controller that performs the following tasks:

- logs the API Object name, namespace and the value of spec.message field.
- copies spec.message field content into status.specEcho field.
- associates a pod for each Dummy API Object that runs nginx.
- copies the status of Pod into status.podStatus field.
- deletes its Pod when Dummy is deleted.

## **Development and Test Environment** ##

- linux mint 21.1
- git (2.34.1)
- docker (23.0.6)
- kubectl (Client v1.27.2)
- minikube (1.30.1)
- go (go1.20.4 linux/amd64)
- operator-sdk (v1.28.1)
- vscode (1.78.2)

## **Source Code** ##

Source code of dummy-operator is available in GitHub:
- https://github.com/iduranli/dummy-operator.git

## **Docker Image** ##
Docker image of dummy-operator is available in Docker Hub:
- docker pull iduranli/dummy-operator:v0.0.1

# **Testing dummy-operator** #

Three terminals will be used to conduct test:
- terminal-1 : for running minikube
- terminal-2 : for running the Dummy controller and observing the logs
- terminal-3 : for interacting with kubernetes environment and observing results

Dummy operator can be tested via an image from Docker Hub or by running directly from the source code. Both testing methods are described below. They both test the same requirements with the same expected results, however test executions are slightly different.

Note: lines starting with "->" are command line executions.

## **1. Testing using image from Docker Hub** ##

**Test flows:**

1. (terminal-1) Start minikube
    - -> minikube start --driver=docker
2. (terminal-2) pull docker image
    - -> docker pull  iduranli/dummy-operator:v0.0.1
3. (terminal-2) service account "default" requires permissions to access objects within kubernetes. run the following yaml file to give access permissions
    - -> curl -O https://raw.githubusercontent.com/iduranli/dummy-operator/main/config/samples/service-account-patch-for-default.yaml
    - -> kubectl apply -f service-account-patch-for-default.yaml
4. (terminal-2) verify the permissions
    - -> kubectl auth can-i --as=system:serviceaccount:default:default list pod
    - result sould be yes
5. (terminal-2) run the custom controller Dummy (wait for the initialization to complete)
    - -> kubectl run dummy-operator --image iduranli/dummy-operator:v0.0.1
6. (terminal-2) observe further logs
    - -> kubectl logs -f dummy-operator -c dummy-operator
    - "object name", "object namespace" and "spec.message" values are printed to screen
    - specEcho is updated with the value of "spec.message"
    - a new Pod is created
    - podStatus is "Pending" and after a while is "Running"
7. (terminal-3) verify specEcho and podStatus field values in kubernetes
    - -> kubectl get dummies.interview.com dummy-sample -o yaml
    - specEcho is updated with value of "spec.message"
    - podStatus is "Running"
8. (terminal-3) verify a new Pod is created in kubernetes
    - -> kubectl get all
    - a Pod exists with name dummy-nginx (and dummy-operator)
9. (terminal-3) modify spec.message field in _v1alpha1_dummy.yaml and apply
    - -> curl -O https://raw.githubusercontent.com/iduranli/dummy-operator/main/config/samples/_v1alpha1_dummy.yaml
    - -> kubectl apply -f _v1alpha1_dummy.yaml
10. (terminal-2) verify results in log
    - observe log for "object name", "object namespace" and "spec.message" values,
11. (terminal-3) verify specEcho and podStatus field values in kubernetes
    - -> kubectl get dummies.interview.com dummy-sample -o yaml
    - specEcho is updated with the value of "spec.message"
    - podStatus is "Running"
12. (terminal-3) delete the pod
    - -> kubectl delete pod dummy-nginx
13. (terminal-2) verify results
    - Pod deletion is detected by controller
    - a new Pod is created
    - podStatus is "Pending" and after a while is "Running"
14. (terminal-3) verify a new Pod is created in kubernetes
    - -> kubectl get all
    - a Pod exists with name dummy-nginx
15. (terminal-3) delete Dummy object
      kubectl delete -f _v1alpha1_dummy.yaml
16. (terminal-2) verify results
    - Pod is deleted by the controller
17. (terminal-3) verify that Pod does not exist in kubernetes
    - -> kubectl get all
    - no Pod exists with name dummy-nginx

## **2. Testing using source code from GitHub** ##

**Test flows:**
1. (terminal-1) Start minikube
    - -> minikube start --driver=docker
2. (terminal-2) get the respository dummy-operator from github
    - -> git clone https://github.com/iduranli/dummy-operator.git
3. (terminal2 & terminal-3) set test directory
    - -> cd dummy-operator
4. (terminal-2) run the custom controller Dummy (wait for the initialization to complete)
    - -> make install run
5. (terminal-3) apply the Dummy object
    - -> kubectl apply -f config/samples/_v1alpha1_dummy.yaml
6. (terminal-2) verify results
    - "object name", "object namespace" and "spec.message" values are printed to screen
    - specEcho is updated with the value of "spec.message"
    - a new Pod is created
    - podStatus is "Pending" and after a while is "Running"
7. (terminal-3) verify specEcho and podStatus field values in kubernetes
    - -> kubectl get dummies.interview.com dummy-sample -o yaml
    - specEcho is updated with value of "spec.message"
    - podStatus is "Running"
8. (terminal-3) verify a new Pod is created in kubernetes
    - -> kubectl get all
    - a Pod exists with name dummy-nginx
9. (terminal-3) modify spec.message field in config/samples/_v1alpha1_dummy.yaml and apply
    - -> kubectl apply -f config/samples/_v1alpha1_dummy.yaml
10. (terminal-2) verify results in log
    - observe log for "object name", "object namespace" and "spec.message" values,
11. (terminal-3) verify specEcho and podStatus field values in kubernetes
    - -> kubectl get dummies.interview.com dummy-sample -o yaml
    - specEcho is updated with the value of "spec.message"
    - podStatus is "Running"
12. (terminal-3) delete the pod
    - -> kubectl delete pod dummy-nginx
13. (terminal-2) verify results
    - Pod deletion is detected by controller
    - a new Pod is created
    - podStatus is "Pending" and after a while is "Running"
14. (terminal-3) verify a new Pod is created in kubernetes
    - -> kubectl get all
    - a Pod exists with name dummy-nginx
15. (terminal-3) delete Dummy object
      kubectl delete -f config/samples/_v1alpha1_dummy.yaml
16. (terminal-2) verify results
    - Pod is deleted by the controller
17. (terminal-3) verify that Pod does not exist in kubernetes
    - -> kubectl get all
    - no Pod exists with name dummy-nginx
