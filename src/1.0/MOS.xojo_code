#tag Module
Protected Module MOS
	#tag Method, Flags = &h0
		Function fromUCS2(s as string) As string
		  // here we assume that UCS-2 Big Endian can be represented with UTF-16 
		  // and convert it back to UTF-8 to maintain XML and web service standards
		  
		  s = DefineEncoding(s,Encodings.UTF16BE)
		  
		  s = ConvertEncoding(s,Encodings.UTF8)
		  
		  return s
		End Function
	#tag EndMethod

	#tag Method, Flags = &h1
		Protected Sub mosHeartbeat()
		  
		End Sub
	#tag EndMethod

	#tag Method, Flags = &h1
		Protected Function mosKeepAlive() As boolean
		  // ABOUT THIS METHOD:
		  // Firewalls often close connections after short periods without traffic. 
		  // The keepAlive message is utilized as a mechanism to keep the connection active, 
		  // especially when MOS passive mode is in use.
		  
		  // <mos>
		  // <mosID>mediaserver.mos</mosID>
		  // <ncsID>newsroomsystem.mos</ncsID>
		  // <keepAlive/>
		  // </mos>
		  
		  dim s as string
		  
		  
		  
		  return true
		End Function
	#tag EndMethod

	#tag Method, Flags = &h1
		Protected Sub mosListMachInfo()
		  
		End Sub
	#tag EndMethod

	#tag Method, Flags = &h1
		Protected Sub mosReqMachInfo()
		  
		End Sub
	#tag EndMethod

	#tag Method, Flags = &h1
		Protected Function mosTime() As string
		  dim d as Xojo.Core.Date = Xojo.Core.Date.now
		  dim s as string
		  
		  // Timestamp format according to MOS Protocol 4.0:
		  // 2009-04-11T14:22:07,125-02:00 or
		  // 2009-04-11T14:22:07,125Z when in UTC time
		  
		  if app.mUTCTimeFormat then
		    s = s + "125Z"
		  else
		    s = s + "125-02:00"
		  end
		  
		End Function
	#tag EndMethod


End Module
#tag EndModule
