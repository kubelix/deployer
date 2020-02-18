# kubelix deployer

The kubelix deployer is an operator managing application deployments on a kubernetes cluster.
It reads the state to enforce from a CRD which looks as follows:

```yaml
apiVersion: apps.kubelix.io/v1alpha1
kind: Service
metadata:
  name: example
spec:
  # singleton=true implies replica = 1 & deploymentStrategy = recreate. Use this for services 
  # where you want to have exactly 1 instance of, or at most 1 instance in case of deployment rollout 
  singleton: false
  image: paulbouwer/hello-kubernetes:1.5
  
  # ports can contain 0 to n ports exposed on a corev1/service
  # if ports is an empty list no service is created at all
  ports:
    - name: http # unique name within the port list
      container: 8080 # port the container exposes
      service: 80 # port the service exposes; also used for ingress

      # each ingress line 
      ingresses:
        - host: "example.klinkert.io"
          paths: ["/"] # a single path with "/" is the default

  # resources of container. Can be left blank.
  resources:
    limits:
      cpu: "1"
      memory: 1Gi
    requests:
      cpu: 100m
      memory: 128Mi

  # environment variables for the application container
  env:
    KEY1: VALUE1
    KEY2: value2
  
  # each file will be mounted at the specified path with the specified content
  files:
    - name: config
      path: /config.yaml
      content: |
        key: value
        map:
          something: different
```

From this service specification the following objects would be created and managed:

- `appsv1/deployment` with the specified container, environment variables, config files and resources
- `corev1/service` with the ports
- `corev1/configMap` with the config files specified
- `networkingv1beta1/ingress` for each ingress specs on the ports


## docker image

The docker image is automatically build and published at https://hub.docker.com/r/kubelix/deployer .
You can either use the latest tag or a specific git tag.


## helm chart

There is a helm chart which is hosted at github pages:

```bash
helm repo add kubelix https://kubelix.github.io/helm-charts/
helm search repo deployer
helm install kubelix/deployer
```


## Assumptions / usage

- Each service only consists of a single container
- Each service has one or more ports
    - each port may have an ingress config
        - each ingress config may have one or more hosts, but paths are configured per host
- Configuration of services is either done with
    - environment variables
    - config files
    - CLI args
- If you need to replace variables in the service custom resource


## Private docker registries

The config file of the deployer contains a section for docker login credentials to be added to all deployments managed by
the operator:

```yaml
dockerPullSecretes:
  - registry: gitlab.com
    username: test-user
    password: test-password
```

> **hint:** This assumes that you have one deployment user configured in your registry that is used for all projects to pull images.
If you need multiple users / credentials the safest way would be to deploy multiple deployer (which then only watch a
single namespace) and thus separate the credentials, because the deployer pragmatically adds all configured docker pull secrets to all
managed services. 


## TODO

- [ ] Liveness & Readiness probes
- [ ] cleanup objects where the related config is removed
        (i.e. by removing all of the ports the service and ingresses should be removed)
- [ ] 
