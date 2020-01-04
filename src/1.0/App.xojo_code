#tag Class
Protected Class App
Inherits ConsoleApplication
	#tag Event
		Function Run(args() as String) As Integer
		  do
		  loop until RegisterPlugins
		  
		  mMyApplication = New MyApplication
		  
		  mMyApplication.Initialize( args )
		  
		  // expected arguments:
		  // mosID - ID of MOS server instance
		  
		  if ubound(args)>0 then
		    print "Starting new server instance with mosID "+args(1)
		    mosID = args(1)
		  else
		    print "Cannot operate without mosID, shutting down."
		    Return mMyApplication.Finalize
		  end
		  
		  // starting server
		  
		  
		  
		  Do
		    DoEvents(100)
		    // Sit in a loop until Idle returns true
		  Loop Until mMyApplication.Idle
		  
		  Return mMyApplication.Finalize
		End Function
	#tag EndEvent


	#tag Method, Flags = &h0
		Function SendToAll(msg as string) As boolean
		  
		  
		  return true
		End Function
	#tag EndMethod

	#tag Method, Flags = &h0
		Function SendToNCS(msg as string, ncsid as integer) As boolean
		  
		  
		  return true
		End Function
	#tag EndMethod


	#tag Property, Flags = &h0
		mCurrentSocket As Integer
	#tag EndProperty

	#tag Property, Flags = &h21
		Private mMyApplication As MyApplication
	#tag EndProperty

	#tag Property, Flags = &h0
		mosID As Variant
	#tag EndProperty

	#tag Property, Flags = &h0
		mUTCTimeFormat As Boolean
	#tag EndProperty


	#tag ViewBehavior
		#tag ViewProperty
			Name="mCurrentSocket"
			Visible=false
			Group="Behavior"
			InitialValue=""
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="mUTCTimeFormat"
			Visible=false
			Group="Behavior"
			InitialValue=""
			Type="Boolean"
			EditorType=""
		#tag EndViewProperty
	#tag EndViewBehavior
End Class
#tag EndClass
