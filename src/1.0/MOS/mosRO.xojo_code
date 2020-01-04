#tag Class
Protected Class mosRO
	#tag Method, Flags = &h0
		Function appendRO(mosItem as MOS.mosROItem) As enumRoStatus
		  
		  
		  
		  return enumRoStatus.OK
		End Function
	#tag EndMethod

	#tag Method, Flags = &h0
		Sub init()
		  
		End Sub
	#tag EndMethod

	#tag Method, Flags = &h0
		Function rowCount() As integer
		  if Tree<>nil then
		    
		    return tree.Count
		    
		  end
		  
		  return 0
		End Function
	#tag EndMethod


	#tag Property, Flags = &h0
		roID As string
	#tag EndProperty

	#tag Property, Flags = &h0
		Tree As AVLTree
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
			Name="Tree"
			Visible=false
			Group="Behavior"
			InitialValue=""
			Type="Integer"
			EditorType=""
		#tag EndViewProperty
	#tag EndViewBehavior
End Class
#tag EndClass
