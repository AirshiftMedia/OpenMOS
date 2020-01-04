# Open Media Object Server
An implementation of Media Object Server using MOS Protocol 4.0. Support for legacy 2.8.2 might be added later.
The project will aim at compliance with Profile 7.

Implementation status:
* [x]  Core
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
This project is not affiliated with MOS Group.

When the maturity level reaches early beta, the project shall make available native binaries for Windows, macOS and Linux. Dependencies are to be kept as minimal as possible.

## Repository Folder Structure
- /bin contains the built application binaries
- /doc contains the technical documentation
- /res contains the additional resources such as image files
- /src contains the source code files in xojo_project -format
