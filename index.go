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

type person struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Age    int    `json:"age"`
}

type addPersonAction struct {
	Group  string
	Person *person `json:"omitempty"`
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

// func decodeActionFromJson(req *http.Request) AddPersonAction {
// 	decoder := json.NewDecoder(req.Body)
// 	var t AddPersonAction
// 	err := decoder.Decode(&t)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return t
// }

func handleAddObject(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var action addPersonAction
	decodeError := decoder.Decode(&action)

	if decodeError != nil || len(action.Group) == 0 ||
		action.Person == nil || len(action.Person.Name) == 0 {
		configureFailureResponse(&w, "Invalid input format")
		return
	}

	groupName := action.Group
	if _, err := os.Stat(groupName); err != nil {
		groupsFile, err := os.OpenFile("groups", os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			_, err2 := groupsFile.WriteString(groupName + ", ")
			if err != nil {
				fmt.Println(err2)
			}
			groupsFile.Sync()
			groupsFile.Close()
		} else {
			configureFailureResponse(&w, "Interal Server Error")
			return
		}
	}

	var groupMap map[string]person
	byteValue, _ := ioutil.ReadFile(groupName)
	err := json.Unmarshal(byteValue, &groupMap)
	if err != nil {
		groupMap = make(map[string]person)
	} else {
		configureFailureResponse(&w, "Interal Server Error")
		return
	}
	p := action.Person
	groupMap[action.Person.Name] = *p
	personString, _ := json.Marshal(groupMap)
	ioutil.WriteFile(groupName, personString, 0644)

	configureSuccessResponse(&w, action.Person.Name+" added to group "+groupName)
	return
}

func handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	// Delete the group if it was the last object in the group
	decoder := json.NewDecoder(r.Body)
	var action deletePersonAction
	decodeError := decoder.Decode(&action)
	groupName := action.Group

	if decodeError != nil || len(action.Group) == 0 ||
		len(action.Person) == 0 {
		configureFailureResponse(&w, "Invalid input format")
		return
	}

	if _, err := os.Stat(groupName); err != nil {
		var f2map map[string]person
		byteValue, _ := ioutil.ReadFile(groupName)
		err := json.Unmarshal(byteValue, &f2map)
		if err != nil {
			f2map = make(map[string]person)
		}
		// read map
		delete(f2map, action.Person)
		personString, _ := json.Marshal(f2map)
		ioutil.WriteFile(groupName, personString, 0644)
	} else {
		configureFailureResponse(&w, "Internal Server Error")
		return
	}
	configureSuccessResponse(&w, "Deleted "+action.Person+" from "+action.Group)
}

func handleGetAllObjects(w http.ResponseWriter, r *http.Request) {

}

func handleDeleteGroup(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// Check if groups file exist and create it if it doesn't
	if _, err := os.Stat("groups"); err != nil {
		os.Create("groups")
	}

	http.HandleFunc("/add", handleAddObject)
	http.HandleFunc("/get", handleGetAllObjects)
	http.HandleFunc("/delete", handleDeleteObject)
	http.HandleFunc("/delete-group", handleDeleteGroup)
	http.ListenAndServe(":8080", nil)
}

func configureSuccessResponse(w *http.ResponseWriter, message string) {
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

func configureFailureResponse(w *http.ResponseWriter, message string) {
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
