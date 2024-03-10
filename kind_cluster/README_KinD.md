# Deploy a KinD cluster

These instructions assume you already have [KinD](https://kind.sigs.k8s.io/) installed and working!

## Steps

The following steps outline creating a KinD cluster and deploying an instance of [nginx](https://www.nginx.com/) just
to test initial connectivity.

Create the cluster:

```bash
prompt> kind create cluster --config k8s_kind_cluster.yml --name figure-test
```

This cluster is setup to map _port 30000_ on the local machine to _port 30000_ on the [Docker](https://docker.com)
container that is hosting the cluster control-plane. The [nginx](https://www.nginx.com/) _Deployment_ manifest includes
a _Service_ definition of type _NodePort_ which maps the same _port 30000_. Finally, the _Service_ routes traffic from
_port 30000_ to _port 80_ inside the _Pod_ which is where nginx is listering.

Deploy nginx:

```bash
prompt> kubectl apply -f k8s_kind_nginx_deploy.yml
```

Verify that nginx has been deployed:

```bash
prompt> kubectl get pods
NAME                          READY   STATUS    RESTARTS   AGE
nginx-test-6f4b69fb78-pt2j2   1/1     Running   0          133m
```

Also:

```bash
prompt> curl localhost:30000

<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

Additional configuration required to access the control-plane from the host or to access additional ports that may be
exposed by deployments is beyond the scope of this setup. Ideally, you may want to deploy an Ingress controller or a
Load Balancer configuration.

---
