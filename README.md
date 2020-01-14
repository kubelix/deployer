# kubelix deployer

## TODO

- [ ] File mounting
- [ ] Store checksums in service status to prevent unnecessary updates  


## Assumptions / usage

- Each project is an isolated application
    - if apps need to communicate with each other the either call their (public) exposed endpoints of use queues
    - each project is an isolated namespace with a network policy denying all pod traffic
- Each service has one or more ports
    - each port may have an ingress config
        - each ingress config may have one or more hosts, but paths are configured per host
