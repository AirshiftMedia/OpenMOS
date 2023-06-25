# Open Media Object Server
An implementation of Media Object Server using MOS Protocol 4.0 with WebSocket as a communication method.
The project will aim at compliance with Profile 7.

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

When the maturity level reaches early beta, the project shall make available a Docker image. Dependencies are to be kept as minimal as possible, using frameworks that are still maintained and active.

## Repository Folder Structure
- /bin contains the locally built application binaries
- /pkg contains the local Go packages
- /src contains the Go source files, readme.md explains the folder and package structure
- /doc contains the technical documentation
- /res contains the additional resources such as image files

## Latest news

30/5/2023   Reworking codebase from Xojo to Go because of performance reasons, will take a while
