package main

import (
	// "bytes"
	// "regexp"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const defaultErrorString = "Internal Server Error"

type person struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Age    int    `json:"age"`
}

type addPersonAction struct {
	Group  string
	Person *person
	// Using pointer as to make this value null if not provided in input.
}

type getAction struct {
	Group string
}

type deleteGroupAction struct {
	Group string
}

type deletePersonAction struct {
	Group  string `json:"group"`
	Person string
}

type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func handleAddObject(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var action addPersonAction
	decodeError := decoder.Decode(&action)

	if decodeError != nil || len(action.Group) == 0 ||
		action.Person == nil || len(action.Person.Name) == 0 {
		sendFailureResponse(&w, "Invalid input format")
		return
	}

	groupName := action.Group
	groupsFile, groupFileError := os.OpenFile("groups", os.O_APPEND|os.O_WRONLY, 0644)
	if groupFileError == nil {
		groupsFile.WriteString(groupName + ", ")
		groupsFile.Sync()
		groupsFile.Close()
	} else {
		sendFailureResponse(&w, defaultErrorString)
		return
	}

	var groupMap map[string]person
	byteValue, _ := ioutil.ReadFile(groupName)
	err := json.Unmarshal(byteValue, &groupMap)
	if err == nil {
		if len(groupMap) == 0 {
			groupMap = make(map[string]person)
		}
	} else {
		sendFailureResponse(&w, "Interal Server Error")
		return
	}

	p := action.Person
	groupMap[action.Person.Name] = *p
	personString, _ := json.Marshal(groupMap)
	ioutil.WriteFile(groupName, personString, 0644)

	sendSuccessResponse(&w, action.Person.Name+" added to group "+groupName)
	return
}

func handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var action deletePersonAction
	decodeError := decoder.Decode(&action)
	groupName := action.Group

	if decodeError != nil || len(action.Group) == 0 ||
		len(action.Person) == 0 {
		sendFailureResponse(&w, "Invalid input format")
		return
	}

	if _, err := os.Stat(groupName); err == nil {
		var f2map map[string]person
		byteValue, _ := ioutil.ReadFile(groupName)
		err := json.Unmarshal(byteValue, &f2map)
		if err == nil {
			f2map = make(map[string]person)
		}
		// read map
		delete(f2map, action.Person)
		personString, _ := json.Marshal(f2map)
		ioutil.WriteFile(groupName, personString, 0644)
	} else {
		sendFailureResponse(&w, defaultErrorString)
		return
	}

	// TODO: Delete the group if it was the last object in the group

	sendSuccessResponse(&w, "Deleted "+action.Person+" from "+action.Group)
}

func handleGetAllObjects(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var action getAction
	decodeError := decoder.Decode(&action)
	groupName := action.Group

	if decodeError != nil || len(action.Group) == 0 {
		sendFailureResponse(&w, "No group specified")
		return
	}

	if _, err := os.Stat(groupName); err == nil {
		var f2map map[string]person
		byteValue, _ := ioutil.ReadFile(groupName)
		json.Unmarshal(byteValue, &f2map)
		personString, _ := json.Marshal(f2map)
		sendSuccessResponse(&w, string(personString))
	} else {
		fmt.Println(err)
		sendFailureResponse(&w, "Group does not exist")
		return
	}
}

func handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var action deletePersonAction
	decodeError := decoder.Decode(&action)
	groupName := action.Group

	if decodeError != nil || len(action.Group) == 0 {
		sendFailureResponse(&w, "No group specified.")
		return
	}

	err := os.Remove(groupName)

	if err == nil {
		sendSuccessResponse(&w, "Successfully deleted "+groupName)
	} else {
		sendFailureResponse(&w, defaultErrorString)
	}
}

func main() {
	// Check if groups file exist and create it if it doesn't
	if _, err := os.Stat("groups.json"); err != nil {
		os.Create("groups")
	}

	http.HandleFunc("/add", handleAddObject)
	http.HandleFunc("/get", handleGetAllObjects)
	http.HandleFunc("/delete", handleDeleteObject)
	http.HandleFunc("/delete-group", handleDeleteGroup)
	http.ListenAndServe(":8080", nil)
}

// Helper Function to send a HTTP Success Response
func sendSuccessResponse(w *http.ResponseWriter, message string) {
	response := apiResponse{"success", message}
	res, err := json.Marshal(response)

	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}

	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(200)
	(*w).Write(res)
}

// Helper Function to send a HTTP Failure Response
func sendFailureResponse(w *http.ResponseWriter, message string) {
	response := apiResponse{"failure", message}
	res, err := json.Marshal(response)

	if err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
		return
	}

	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(401)
	(*w).Write(res)
}
