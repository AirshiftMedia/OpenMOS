package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func initAPI() {

	router := newRouter()

}

func main() {

	// check if arguments were provided when invoked

	for _, arg := range os.Args {
		fmt.Println("Arguments provided for initialization: ", arg)
	}

	//

	fmt.Println("Starting OpenMOS service: Now listening on port 8081")

	//


}
