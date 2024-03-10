# Welcome

Welcome to Figure's DevOps skills assessment!

The goal of this assessment is to get an idea of how you work and your ability to speak in depth about the details in your work. Generally, this assessment should not take you longer than 30 minutes to complete. 

Your answers will be reviewed with you in a subsequent interview.

## Instructions

1. Click on the green "Use This Template" button in the upper-right corner and create a copy of this repository in your own GitHub account.
2. Name your respository and ensure that it's public, as you will need to share it with us for review.
3. When you have completed the questions, please send the URL to the recruiter.

## Assessments

### Kubernetes

1. Fix the issues with this Kubernetes manifest to ensure it is ready for deployment.
2. Add the following limits and requests to the manifest:
- CPU limit of 0.5 CPU cores
- Memory limit of 256 Mebibytes
- CPU request of 0.2 CPU cores
- Memory request of 128 Mebibytes

>__NOTE:__ Corrected manifest appears below. Also included as [k8s_nginx_deploy_fixed.yml](k8s_nginx_deploy_fixed.yml).

```yaml
apiVersion: apps/v1
# Correct `kind` value "Deploy" -> "Deployment"
kind: Deployment
metadata:
  name: nginx-deploy
  labels:
    app: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        # Correct image tag: "nginx:current" -> "nginx:latest"
        image: nginx:latest
        ports:
        - containerPort: 80
        # Set resource limits to:
        # - cpu: 0.5 CPU cores
        # - memory: 256 Mebibytes
        # Set request limits to:
        # - cpu: 0.2 CPU cores
        # - memory: 128 Mebibytes
        #
        # NOTE: If you set the limit higher than the request, you will be subject to pod eviction if the request exceeds
        # the defined limit. The request only represents the initial (or minimum) amount of memory that the pod needs in
        # order to be scheduled on a node.
        #
        # For cpu, it is generally inadvisable to set a limit in most cases because it will throttle your application in
        # times when adequate cpu headroom is available. The cpu request is used by the scheduler to job placement.
        resources:
          limits:
            cpu: "500m"    # 0.5 CPU cores
            memory: "256Mi"  # 256 Mebibytes
          requests:
            cpu: "200m"    # 0.2 CPU cores
            memory: "128Mi"  # 128 Mebibytes
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
  type: ClusterIP
  ```

### Go

Write a script in Go that redeploys all pods in a Kubernetes cluster that have the word `database` in the name.

Requirements:
- Assume local credentials in your kube config have full access. There is no need to connect via a service account, etc.
- You must use the [client-go](https://github.com/kubernetes/client-go) library.
- Your script must perform a graceful restart, similar to kubectl rollout restart. Do not just delete pods.
- You must use Go modules (no vendor directory).

>__NOTE:__ You should be able to run the app with `go mod init` followed by `go run .` in the root directory.

#### More Notes

The app will create two Deployments containing a total of 4 Pods (two Deployments with two replicas for each). The
deployed app is [Postgres]() but for the purposes of this assignment it could be anything. Deployments and Pods have
the name "database" in them. The app can be run multiple times, the Deployments will only be created once.

The app will query the cluster for Pods that have names which contain the word "database."

It's important to note that the app obtains a list of Pods that match the search criteria and then queries for the
Deployments (via ReplicaSets) that contain the Pods. It was done this way to accomodate performing a brutal replacement
of ad-hoc Pods in the future. That is, you can only gracefully restart Pods that are part of ReplicaSets because it is
the ReplicaSet that handles the creation of new Pods before the old ones are terminated.

The way the __graceful restart__ is handled is that an Annotation is appended/updated on the target Deployments. This
seemingly innocuous change is enough to tell the backend to rollout new Pods. The annotation looks like this:

```text
Annotations:      force/deploy: 2024-03-10 21:40:01.664627499 +0000 UTC
```

The value is the timestamp for the re-deploy.

#### Final Comments

I know this could be handled in a 30m exercise but I had fun with this and was exploring the use of KinD while also
setting up the toolchain on a new computer. What it means is you will see there is a `k8s` package that I created in
here with a bunch of functionality that's not used for this assignment but just is more of an example of the way I
develop a codebase. Obviously, for the real thing I would remove all uncalled code as it just makes long term
maintainability more complicated (i.e. is adds "noise").

Also, for absolute narrow scope, I would refactor to just grab Deployments from the start, retrieve references to their
Pods, filter those by name, and "kick" the Deployments with Pods that match the filter criteria.

The KinD cluster configuration that I used here is dropped in as well along with a brief [README](kind_cluster/README_KinD.md)
file that explains setup.

---
