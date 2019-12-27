#tag Class
Protected Class MyApplication
	#tag Method, Flags = &h0
		Function Finalize() As Integer
		  Print "Shutting down..."
		  
		  Return 0
		End Function
	#tag EndMethod

	#tag Method, Flags = &h0
		Function getProcessStatus() As string
		  dim s as string
		  
		  if TargetMacOS then
		    s = "Memory " +format(SystemInformationMBS.AvailableRAM/1024/1024,"0")+" MB of RAM free."
		  elseif TargetWindows then
		    s = "Memory " +format(SystemInformationMBS.AvailableRAM/1024/1024,"0")+" MB of RAM free."
		  elseif TargetLinux then
		    s = "Memory " +format(SystemInformationMBS.AvailableRAM/1024/1024,"0")+" MB of RAM free."
		  end
		  
		  
		  return s
		End Function
	#tag EndMethod

	#tag Method, Flags = &h0
		Function Idle() As Boolean
		  mLoopSystemCheck = mLoopSystemCheck+ 1
		  mLoopKeepAlive = mLoopKeepAlive + 1
		  
		  
		  if mLoopKeepAlive >= 300 then // recommended not to be less than 30 sec
		    
		    do
		    loop until MOS.mosKeepAlive
		    
		    mLoopKeepAlive = 0
		    
		  end
		  
		  if mLoopSystemCheck >= 300 then
		    
		    
		    print "System status: " + getProcessStatus
		    mLoopSystemCheck = 0
		    
		  end
		  
		  
		  return false
		  
		End Function
	#tag EndMethod

	#tag Method, Flags = &h0
		Sub Initialize(args() as String)
		  Print "Initializing the application"
		End Sub
	#tag EndMethod


	#tag Property, Flags = &h0
		mLoopKeepAlive As Integer
	#tag EndProperty

	#tag Property, Flags = &h0
		mLoopSystemCheck As Integer
	#tag EndProperty


	#tag ViewBehavior
		#tag ViewProperty
			Name="Name"
			Visible=true
			Group="ID"
			InitialValue=""
			Type="String"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="Index"
			Visible=true
			Group="ID"
			InitialValue="-2147483648"
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="Super"
			Visible=true
			Group="ID"
			InitialValue=""
			Type="String"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="Left"
			Visible=true
			Group="Position"
			InitialValue="0"
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="Top"
			Visible=true
			Group="Position"
			InitialValue="0"
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
		#tag ViewProperty
			Name="mLoopSystemCheck"
			Visible=false
			Group="Behavior"
			InitialValue=""
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
	#tag EndViewBehavior
End Class
#tag EndClass
