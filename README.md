# Open Media Object Server
An implementation of Media Object Server using MOS Protocol 4.0 with WebSocket as a communication method.
The project will aim at compliance with Profile 7.
> [!NOTE]
> At initial stages, message attributes differ from the protocol specification due to practical reasons.

Implementation status:
* [ ]  Core
* [ ]  Profile 0
* [ ]  Profile 1
* [ ]  Profile 2
* [ ]  Profile 3
* [ ]  Profile 4
* [ ]  Profile 5
* [ ]  Profile 6
* [ ]  Profile 7

As an experimental feature, the Profile 5 roCtrl will be implemented in a way that it can support IoT device
control using MQTT protocol. An example use case is the red light control. In the future this could be expanded to actual machine controls with protocols like Ember+ (https://github.com/Lawo/ember-plus) or VDCP.

For more information about the MOS protocol, see https://www.mosprotocol.com
This project is not affiliated with MOS Group and will use the word compliance according to the requirements set by the MOS Group.

Logging system utilizes Sentry observability layer. Developer subscription is available for free at https://www.sentry.io

When the maturity level reaches early beta, the project shall make available a Docker image. Dependencies are to be kept as minimal as possible, using frameworks that are still maintained and active.

![MOS Project Official Logo](/res/mosproject-logo.jpg)
