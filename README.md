# OpenMOS Media Object Server

![MOS Project Official Logo](/res/mosproject-logo.jpg)

An implementation of Media Object Server using MOS Protocol 4.0 with WebSocket as a communication method.
The project will aim at compliance with Profile 7.
> [!NOTE]
> At initial stages, message attributes differ from the protocol specification due to practical reasons.

Implementation status:
* [x]  Core
* [x]  MongoDB Data Repository
* [x]  Sentry Observability
* [ ]  Profile 0 - Basic Communication
* [ ]  Profile 1 - Basic Object Based Workflow
* [ ]  Profile 2 - Basic Running Order / Content List Workflow
* [ ]  Profile 3 - Advanced Object Based Workflow
* [ ]  Profile 4 - Advanced RO/Content List Workflow
* [ ]  Profile 5 - Item Control
* [ ]  Profile 6 - MOS Redirection
* [ ]  Profile 7 - MOS RO/Content List Modification

As an experimental feature, the Profile 5 roCtrl will be implemented in a way that it can support IoT device
control using MQTT protocol. An example use case is the red light control. In the future this could be expanded to actual machine controls with protocols like Ember+ (https://github.com/Lawo/ember-plus) or VDCP.

For more information about the MOS protocol, see https://www.mosprotocol.com
This project is not affiliated with MOS Group and will use the word compliance according to the requirements set by the MOS Group.

Logging system utilizes Sentry observability layer. Developer subscription is available for free at https://www.sentry.io

When the maturity level reaches early beta, the project shall make available a Docker image. Dependencies are to be kept as minimal as possible, using frameworks that are still maintained and active.

### Configuration file:

```
# Application configuration
app:
  name: mos.station.com
  version: 1.0.0
  environment: development

# MOS service configuration
server:
  host: 0.0.0.0
  port: 10540
  readTimeout: 5s
  writeTimeout: 5s
  shutdownTimeout: 30s
  ```

Generate default configuration file:
```
./openmos --generate-config=config.yaml
```

