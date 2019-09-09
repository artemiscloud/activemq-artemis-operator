
## Deployment Configuration for ActiveMQ Artemis broker 
 
#### Deploying an ActiveMQ Artemis broker with Persistent

 Add persistent flag true in the custom resource file:
 
e.g.

```yaml
          persistenceEnabled: true
    
 ```

## Trigger a ActiveMQ Artemis deployment

Use the console to `Create Broker` or create one manually as seen below. Ensure SSL configuration is correct in the
custom resource file.

```bash
$ oc create -f deploy/crs/broker_v2alpha1_activemqartemis_cr.yaml
```

## Clean up an ActiveMQ Artemis deployment

```bash
oc delete -f deploy/crs/broker_v2alpha1_activemqartemis_cr.yaml
```

