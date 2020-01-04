# Field Descriptions

(Source: http://mosprotocol.com/wp-content/MOS-Protocol-Documents/MOSProtocolVersion40/index.html)


| Field | Description |
| -------- | ---------------- |
| b | Bold face type:  Specifies that text between tags is in boldface type. |
| canClose | Indicates whether an NCS can close the window in which it is hosting an ActiveX control when the control sends an ncsReqAppClose message. Permitted values are TRUE and FALSE |
| changed | Changed Time/Date: Time the object was last changed in the MOS. Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC. |
| changedBy | Last Changed by: Name of the person or process that last changed the object in the MOS. This can be stored in a language other than English. |
| command | roItemCtrl command: The commands READY, EXECUTE, PAUSE and STOP, as well as general indicator, SIGNAL, can be addressed at each MOS Structure level.  In other words, a single command can begin EXECUTION of an entire Running Order, of a Story containing multiple Items, or of a single Item. |
| controlDefaultParams | This value represents the parameters that can be passed to an ActiveX. |
| controlFileLocation | controlFileLocation is the file location for the default ActiveX control. |
| controlName | This value represents the key/classid key used to load the ActiveX from the registry. |
| controlSlug | Defined by MOS 128 characters max |
| created | Creation Time/Date: Time the object was created in the MOS. Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC. |
| createdBy | Created by: Name of the person or process that created the object in the MOS. This can be stored in a language other than English. 128 chars max. |
| defaultActiveX | defaultActiveX contains tags that describe the correct settings for the ActiveX control (NOTE: no two <defaultActivX> elements can have the same <mode> value). |
| description | Object Description: Text description of the MOS object. No maximum Length is defined. This can be stored in a language other than English. |
| deviceType | deviceType is a required attribute of supportedProfiles.  The required values are either "NCS" or "MOS". |
| DOM | Date of Manufacture. |
| element_source | element_source is a tag that designates the story(s) and or item(s) to be acted upon. |
| element_target | element_target specifies where in the running order the actions are to take place. |
| em | Emphasized Text: markup within description and p to emphasize text. |
| endx | Used in MOS ActiveX messages. The maximum width in pixels that the NCS Host allows for an ActiveX Plug-In in a particular metric in ncsAppInfo. 0XFFFFFFFF max. |
| endy | Used in MOS ActiveX messages. The maximum height in pixels that the NCS Host allows for an ActiveX Plug-In in a particular metric in ncsAppInfo. 0XFFFFFFFF max. |
| generalSearch | String used for simple searching in the mosReqObjList message. Logical operators are allowed. 128 chars max. |
| hwRev | HW Revision: 128 chars max. |
| I | Italics:  Specifies that text between tags is in Italics. |
| ID | Identification of a Machine: text. 128 chars max. |
| item | Item: Container for item information within a Running Order message. |
| itemChannel | Item Channel: Channel requested by the NCS for MOS to playback a running order item. 128 chars max. |
| itemEdDur | Item Editorial Duration: in number of samples 0XFFFFFFFF max. |
| itemEdStart | Editorial Start: in number of samples 0XFFFFFFFF max. |
| itemID | Item ID: Defined by NCS, UID not required. 128 chars max. |
| itemSlug | Item Slug: Defined by NCS. 128 chars max |
| itemTrigger | Item Air Trigger: "MANUAL", "TIMED" or "CHAINED". |
| CHAINED (sign +/-) (value in # of samples) | CHAINED -10 would start the specified clip 10 samples before the proceeding clip ended. CHAINED 10 would start the specified clip 10 samples after the preceding clip ended, thus making a pause of 10 samples between the clips. There is a space character between the word CHAINED and the value. |            
| itemUserTimingDur | Item User Timing Duration: If the NCS extracts a duration value from a MOS item for show timing, and this field has a value, then the NCS must use this value. The value is in number of samples. 0XFFFFFFFF max. | 
| leaseLock | Interger value used in roReqStoryAction less then 999 where a MOS requests a lock on a given story from the NCS.  A MOS must then send subsequent message to modify the story before the leaseLock expires. | 
| listReturnEnd | Integer value used in mosReqObjList commands to specify the index of the last mosObj message requested or returned in a message. 0xFFFFFFFF max. | 
| listReturnStart | Integer value used in mosReqObjList commands to specify the index of the first mosObj message requested or returned in a message. 0xFFFFFFFF max. | 
| listReturnStatus | Optional string value in the mosObjList message that specifies the reason for a zero value in the <listReturnTotal> tag. 128 chars max. | 
| listReturnTotal | Integer value in the mosObjList message specifying how many mosObj messages are returned. 0xFFFFFFFF max. | 
| mosPlugInID | mosPlugInID is used as the unique identifier for a ActiveX control.  128 characters max. | 
| macroIn | Macro Transition In: Defined by MOS. 128 chars max. | 
| macroOut | Macro Transition Out: Defined by MOS. 128 chars max. | 
| manufacturer | Used in MOS ActiveX messages. Manufacturer: Text description. 128 chars max. |
| messageID | Unique identifier sent with requests. See chapter 4.1.6 for a detailed description.  Format: signed integer 32-bit, value above or equal to 1, decimal or hexadecimal.  An empty messageID tag is allowed for messages when used in the ActiveX interface. |
| mode | Used in MOS ActiveX messages. How the ActiveX Plug-In window appears in the NCS Host window: MODALDIALOG, MODELESS, CONTAINED, TOOLBAR.|
| model | Model: Text description. 128 chars max. |
| modifiedBy | Modified by: Name of the person or process that last modified the object in the MOS. This can be stored in a language other than English. 128 chars max. |
| mosAbstract | Abstract of the Object intended for display by the NCS.  This field may contain HTML and DHTML markup.  The specific contents are limited by the NCS vendor's implementation.  Length is unlimited but reasonable use is suggested. |
| mosActiveXversion | Used in MOS ActiveX messages. String indicating the version of the ActiveX Plug-In. 128 chars max |
| mosID | MOS ID: Character name for the MOS unique within a particular installation. |
| mosGroup | This field is intended to imply the name of a destination, group or folder for the Object pointer to be stored in the NCS.  128 chars max. |
| mosMsg data type | Used in MOS ActiveX messages. Clipboard format used for OLE drag and drop from the ActiveX Plug-In. |
| mosPayload | mosPayload is apart of the mosExternalMetadat block, and it generally includes essential metadata that is referenced within the mosSchema. |
| mosPlugInID | Used in MOS ActiveX messages. ID that the NCS Host can use to instantiate the ActiveX Plug-In. 128 chars max. |
| mosProfile | This field is intended to define a device's supported MOS Profiles.  A "YES" or "NO" value is required for each profile.|
| mosRev | MOS Revision: Text description. 128 chars max. |
| mosSchema | mosSchema is the descriptive schema used within the mosPayload.  The value is to be an implied pointer or URL to the actual schema document. |
| mosScope | This field implies the extent to which the mosExternalMetadata block will move through the NCS workflow. Accepted values are "OBJECT" "STORY" and "PLAYLIST" |
| ncsID | NCS ID: Character name for the NCS unique within a particular installation. 128 chars max. |
| objAir | Air Status: "READY" or "NOT READY". |
| objDur | Object Duration: The number of samples contained in the object. For Still Stores this would be 1. 0XFFFFFFFF MAX |
| objGroup | Definition of the values for objGroup is left to the configuration and agreement of MOS connected equipment.  The intended use is for site configuration of a limited number of locally named storage folders in the NCS or MOS. |
| objID | Object UID: Unique ID generated by the MOS and assigned to this object. 128 chars max. |
| objPath | This is an unambiguous path to a media file - essence.  The field length is 255 chars max, but it is suggested that the length be kept to a minimum number of characters. As of MOS 2.8.5 ObjPaths paths must meet the following requirements:
-Be a call to return the media , without requiring client-side redirection
-The character string following the last slash in the path must be the full filename, including the asset’s extension.These path formats are acceptable:  
-UNC (eg: \\machine\directory\file.extension)
-URL (eg: http://machine/directory/file.extension) http and https
-FTP (eg: ftp://machine/directory/file.extension) ftp and ftps |
| objProxyPath | This is an unambiguous path to a media file – proxy.  The field length is 255 chars max, but it is suggested that the length be kept to a minimum number of characters. As of MOS 2.8.5 ObjPaths paths must meet the following requirements:
Be a call to return the media , without requiring client-side redirection
The character string following the last slash in the path must be the full filename, including the asset’s extension.These path formats are acceptable:
UNC (eg: \\machine\directory\file.extension)

URL (eg: http://machine/directory/file.extension) http and https

(note: FTP is *NOT* allowed for objProxyPath) |
| objMetadataPath | This is an unambiguous path to the xml file – MOS Object.  This field length is 255 chars max, but is is suggested that the length be kept to a minimum number of characters.  These path formats are acceptable:

UNC (eg: \\machine\directory\file.extension)

URL (eg: http://machine/directory/file.extension)

FTP (eg: ftp://machine/directory/file.extension) |
| objRev | Object Revision Number: 999 max. |
| objSlug | Object Slug: Textual object description. 128 chars max. |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |
|  |  |

objTB

Object Time Base: Describes the sampling rate of the object in samples per second. This tag should be populated with a value greater than 0. For PAL Video this would be 50. For NTSC it would be 59.94. For audio it would reflect the audio sampling rate. Object Time Base is used by the NCS to derive duration and other timing information. 0XFFFFFFFF MAX

objType

Object Type: Choices are "STILL", "AUDIO", "VIDEO".

opTime

Operational Time: date and time of last machine start. Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC.

p

Paragraph:  Standard html delimitation for a new paragraph.

pause

Item Delay: Requested delay between items in ms 0XFFFFFFFF MAX.

pi

Presenter instructions:  Instructions to the anchor or presenter that are not to be read such as "Turn to 2-shot."

pkg

            Package:  Specifies that text is verbatim package copy as opposed to copy to be read by presenter.

product

Used in MOS ActiveX messages. String indicating the product name of the NCS Host. 128 chars max.

queryID

Unique identifier used in mosReqObjList and mosObjList to allow the MOS to do cached searching. 128 chars max.

Read1stMEMasBody

Allows the first MEM block to substitute the story body.

ro

The ro tag is used within the roListAll message.  ro designates each individual running order within the roListAll message.

roAir

Air Ready Flag: "READY" or "NOT READY".

roChannel

Running Order Channel: default channel requested by the NCS for MOS to playback a running order. 128 chars max.

roCtrlCmd

Running Order Control Command:  READY, EXECUTE, PAUSE and STOP, as well as general indicator, SIGNAL, can be addressed at each level.  In other words, a single command can begin EXECUTION of an entire Running Order, of a Story containing multiple Items, or of a single Item.

roCtrlTime

Running Order Control Time:  roCtrlTime is an optional field which provides a mechanism to time stamp the time of message transmission, or optionally, to provide a time in the immediate future at which the MOS should execute the command.   Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC.

roEdDur

Running Order Editorial Duration: duration of entire running order. Format in hh:mm:ss, e.g. 00:58:25.

roEdStart       

Running Order Editorial Start: date and time requested by NCS for MOS to start playback of a running order. Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC.

roEventTime

Running Order Event Time: Time of the time cue sent to the parent MOS by the NCS for a specific item event.

roEventType

Running Order Event Type: The type of event that is being queued in the Running order.

roID

Running Order UID: Unique Identifier defined by NCS. 128 chars max.

roSlug

Running Order Slug: Textual Running Order description. 128 chars max.

roStatus

Running Order Status: Options are: "OK" or error description. 128 chars max.

roTrigger

Running Order Air Trigger: "MANUAL", "TIMED" or "CHAINED".

CHAINED (sign +/-) (value in # of samples)

CHAINED -10 would start the specified clip 10 samples before the proceeding clip ended. CHAINED 10 would start the specified clip 10 samples after the preceding clip ended, thus making a pause of 10 samples between the clips. There is a space character between the word CHAINED and the value.slug  Textual Object ID: This is the text slug of the object and is stored in the native language. This can be stored in a language other than English. 128 chars max.

runContext

Used in MOS ActiveX messages. Specifies the context in which the NCS Host is instantiating the ActiveX Plug-In: BROWSE, EDIT, CREATE.

searchField

searchField contains attributes that are originally derived from the schema returned in the initial mosListSearchableSchema message.  

searchGroup

searchGroup contains specific queries based on the values of selected mosExternalMetadata fields.

SN

Serial Number: text serial number. 128 chars max.

startx

Used in MOS ActiveX messages. The minimum width in pixels that the NCS Host allows for an ActiveX Plug-In in a particular metric in ncsAppInfo. 0XFFFFFFFF max.

starty

Used in MOS ActiveX messages. The minimum height in pixels that the NCS Host allows for an ActiveX Plug-In in a particular metric in ncsAppInfo. 0XFFFFFFFF max.

 status

Status: Options are "NEW" "UPDATED" "MOVED" "BUSY " "DELETED", "NCS CTRL", "MANUAL CTRL", "READY", "NOT READY", "PLAY," "STOP".

statusDescription

Status Description: textual description of status. 128 chars max.

 story

Story: Container for story information in a Running Order message.

storyBody

            Story Body: The actual text of the story within a running order.

storyID

Story UID: Defined by the NCS. 128 chars max.

storyItem

Story Item: An item imbedded into a story that can be triggered when that point in the story is reached in the teleprompter.

storyNum

Story Number:  The name or number of the Story as used in the NCS.  This is an optional field originally intended for use by prompters. 128 chars max.

storyPresenter

Story Presenter:  The anchor or presenter of a story within an running order.

storyPresenterRR

Story Presenter Read Rate:  The read rate of the anchor or presenter of a story within a running order.

storySlug

Story Slug: Textual Story description. 128 chars max.

supportedProfiles

This field is intened to determine the device type of the device's supported MOS Profiles.

swRev

Software Revision: (MOS) Text description. 128 chars max.

tab

Tab: tabulation markup within description and p.

techDescription

techDescription is an attribute of objPath and objProxyPath.  This attribute provides a brief and very general technical description fo the codec or file format (NOTE: the codec name should come first, followed by additional optional descriptions).

time

Time: Time object changed status. Format is YYYY-MM-DD'T'hh:mm:ss[,ddd]['Z'], e.g. 2009-04-11T14:22:07,125Z or 2009-04-11T14:22:07,125-05:00.  Parameters displayed within brackets are optional. [,ddd] represents fractional time in which all three digits must be present. ['Z'] indicates time zone which can be expressed as an offset from UTC in hours and minutes.  Optionally, the time zone may be replaced by the character 'Z' to indicate UTC.

u 

Underline: Specifies that text between tags is to be underlined.

userName  

An attribute in the mosReqObjList family of messages and the roReqStoryAction messages which identifies a username

version

Used in MOS ActiveX messages. String indicating the version of the NCS Host. 128 chars max.
