package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

/******************************************************************
*** I added the follwing handler to the server example with     ***
*** a template that handles the html/css and javascript imports ***
*******************************************************************/

// ClientHandler see above
func ClientHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("client.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)
}

func main() {
	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	store := filestore.FileStore{
		Path: "./uploads",
	}

	// A storage backend for tusd may consist of multiple different parts which
	// handle upload creation, locking, termination and so on. The composer is a
	// place where all those separated pieces are joined together. In this example
	// we only use the file store but you may plug in multiple.
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	// Create a new HTTP handler for the tusd server by providing a configuration.
	// The StoreComposer property must be set to allow the handler to function.
	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:              "/files/",
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
	})
	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {
			event := <-handler.CompleteUploads
			fmt.Printf("event: %v\n", event)
			fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}()

	/*********************************************************************************
	 ** The following two lines are all I add to main() from  the tus.io golang     **
	 ** net/http example.                                                           **
	 *********************************************************************************/
	http.Handle("/", http.FileServer(http.Dir("scripts")))
	http.HandleFunc("/client", ClientHandler)
	rh := http.RedirectHandler("/client", 307)
	http.Handle("/client/", rh)

	// Right now, nothing has happened since we need to start the HTTP server on
	// our own. In the end, tusd will start listening on and accept request at
	// http://localhost:8080/files
	http.Handle("/files/", http.StripPrefix("/files/", handler))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(fmt.Errorf("Unable to listen: %s", err))
	}
}
