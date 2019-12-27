#tag Module
Protected Module MOS
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


End Module
#tag EndModule
