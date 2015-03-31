package main

import (
	"fmt"
	"bufio"
	"net/rpc"
	"strings"
	"net/rpc/jsonrpc"
	"os"
	"net"
	"errors"
	"reflect"
	"encoding/json"
	"strconv"
	"log"
)
/*Refers to the integer structure that points to the function that is being called. It is used for
*registering the rpc service.*/
type Dict3 int

type FileType struct{
	File string `json:"file"`
}

/*The refers to the input JSON message structure that represents the configuration details of the server.
 * Client ID: Refers to the client ID.
 * Protocol: Refers to the protocol used to contact the remote server and is mostly TCP.
 * IPAddress: Refers to the IP address of the client.
 * Port: Refers to the port number being used to start the communication process.
 * PersistentStorageContainer: The location of the DICT3 file.
 * Methods: Contains the list of all the functions that can be called at the remote server. */
type config struct{
	serverID string `json:"serverID"`
	Protocol string `json:"protocol"`
	IpAddress string `json:"ipAddress"`
	Port int `json:"port"`
	PersistentStorageContainer FileType `json:"persistentStorageContainer"`
	Methods []string `json:"methods"`
}

/* The JSON message structure of the input passed to the function.
 * Method: The function to be called.
 * Params: The input json object that is passed as an argument to the function.
 * Id: The unique id refers to the transaction between the client and the server.*/
type JsonMessage struct{
	Method string 	`json:"method"`
	Params []interface{} `json:"params"`
	Id int `json:"id"`
}

/* The json result structure that will be displayed to the user after the completion
 * of the insert function.
 * Result: Indicates the successful or unsuccessful completion of the function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the function call. */
type JsonResultInsert struct{
	Result bool `json:"result"`
	Id int `json:"id"`
	Error string `json:"error"`
}

/* The JSON result structure that will be displayed to the user after the completion
 * of the lookup function.
 * Result: Indicates the return of the params identified by the key and the relationship.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonResultLookUp struct{
	Result []interface{} `json:"result"`
	Id int `json:"id"`
	Error string `json:"error"`
}

/* The json result structure that will be displayed to the user after the completion
 * of the listKey function.
 * Result: Identifies the list of unique keys returned by the function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonListKeys struct{
	Result []string `json:"result"`
	Id int `json:"id"`
	Error string `json:"error"`
}


/* The JSON result structure that will be displayed to the user after the completion
 * of the listID function.
 * Result: Identifies the list of unique key and relationship combination returned by the listID function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonListIDs struct{
	Result [][]string `json:"result"`
	Id int `json:"id"`
	Error string `json:"error"`
}

/*The structure of the DICT3 file.
*Key: Refers to the key value of the parameter.
*Relationship: Refers to the relationship value of the parameter. Key and Relationship value together identifies an unique row of the DICT3 file.
*Value: Refers to the value of the argument passed to the function. It is a JSON object. */
type DICT3format struct{
	Key,Relationship string
	Value map[string]interface{}
}

/*This structure is used when there is no need to display any output JSON message to the user. */
type NoOutput struct {
	Error string
}

var serverconfig config
var close bool

/*The lookUp function is used to return the value referred by an existing ID(key + relationship).
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) LookUp(input JsonMessage, output *JsonResultLookUp) error {
	if len(string(input.Params[0].(string))) == 0 || len(string(input.Params[1].(string))) == 0  {
			return errors.New("Key error - Key or Relationship attributes cannot be null in the input")
	}
	key := string(input.Params[0].(string))
	relationship := string(input.Params[1].(string))
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		checkError(err)
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			if key == DICT3exist.Key && relationship == DICT3exist.Relationship {
				output.Result = []interface{}{key,relationship,DICT3exist.Value}
				return nil
			}
			DICT3exist.Value = nil
		}
		return errors.New("Key and/or Relationship not found in DICT3")
	}else{
		return errors.New("DICT3 is empty. Insert values to DICT3 and later look them up")
	}
}

/*The insert function is used to enter an unique triplet in the DICT3 file.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Insert(input *JsonMessage, output *JsonResultInsert) error {
	if len(string(input.Params[0].(string))) == 0 || len(string(input.Params[1].(string))) == 0  {
		return errors.New("Key error - Key or Relationship attributes cannot be null in the input")
	}
	
	if reflect.ValueOf(input.Params[2]).Kind() != reflect.Map {
		return errors.New("Value error - Not   aa valid Json object")
	}
	DICT3input := DICT3format{strings.TrimSpace(string(input.Params[0].(string))),strings.TrimSpace(string(input.Params[1].(string))),map[string]interface{}(input.Params[2].(map[string]interface{}))}
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		defer inFile.Close()
		checkError(err)
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			if DICT3input.Key == DICT3exist.Key && DICT3input.Relationship == DICT3exist.Relationship {
				return errors.New("Key error - Key and Relationship already present in DICT3. Use insertOrUpdate function to change values of existing key.")	
			}
			DICT3exist.Value = nil
		}
	}
	
	saveJSON(serverconfig.PersistentStorageContainer.File,DICT3input)
	output.Result = true
	output.Id = input.Id
	return nil
}

/*The insertOrUpdate function is used to enter an unique triplet in the DICT3 file or update an existing ID(key + relationship) with a new value.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) InsertOrUpdate(input *JsonMessage, output *NoOutput) error {
	if len(string(input.Params[0].(string))) == 0 || len(string(input.Params[1].(string))) == 0  {
		return errors.New("Key error - Key or Relationship attributes cannot be null in the input")
	}
	
	if reflect.ValueOf(input.Params[2]).Kind() != reflect.Map {
		return errors.New("Value error - Not   aa valid Json object")
	}
	
	DICT3input := DICT3format{strings.TrimSpace(string(input.Params[0].(string))),strings.TrimSpace(string(input.Params[1].(string))),map[string]interface{}(input.Params[2].(map[string]interface{}))}
	flag := false
	var lines []string
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		checkError(err)
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			if DICT3input.Key == DICT3exist.Key && DICT3input.Relationship == DICT3exist.Relationship {
				if reflect.DeepEqual(DICT3exist,DICT3input){
						return errors.New("The ID is already set with the same value")
				}
				flag = true
				continue
			}
			lines = append(lines, scanner.Text())
			DICT3exist.Value = nil
		}		
		inFile.Close()
	}
	if flag == false{
		saveJSON(serverconfig.PersistentStorageContainer.File,DICT3input)
	}else{
		os.Remove(serverconfig.PersistentStorageContainer.File)
		outFile, err :=  os.Create(serverconfig.PersistentStorageContainer.File);
		checkError(err)
		w := bufio.NewWriter(outFile)
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		outFile.Close()
		saveJSON(serverconfig.PersistentStorageContainer.File,DICT3input)
	}
	output.Error = " "
	return nil
}

/*The delete function is used to delete a unique triplet from DICT3.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Delete(input *JsonMessage, output *NoOutput) error {
	if len(string(input.Params[0].(string))) == 0 || len(string(input.Params[1].(string))) == 0  {
		return errors.New("Key error - Key or Relationship attributes cannot be null in the input")
	}
	key := strings.TrimSpace(string(input.Params[0].(string)))
	relationship := strings.TrimSpace(string(input.Params[1].(string)))
	flag := false
	var lines []string
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		checkError(err)
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			if key == DICT3exist.Key && relationship == DICT3exist.Relationship {
				flag = true
				continue
			}
			lines = append(lines, scanner.Text())
			DICT3exist.Value = nil
		}		
		inFile.Close()
	}
	if flag == false{
		return errors.New("The Key and Relationship is not present in DICT3")
	}else{
		os.Remove(serverconfig.PersistentStorageContainer.File)
		outFile, err :=  os.Create(serverconfig.PersistentStorageContainer.File);
		checkError(err)
		w := bufio.NewWriter(outFile)
		for _, line := range lines {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		outFile.Close()
	}
	output.Error = " "
	return nil
}

/*The listKey function is used to return a list of unique key values from DICT3.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) ListKeys(input *JsonMessage, output *JsonListKeys) error {
	keymap := make(map[string]struct{})
	key := []string{}
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		checkError(err)
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			keymap[DICT3exist.Key] = struct{}{}
		}
		for k := range(keymap){
			key = append(key,k)
		}
		output.Result = key
		output.Id = input.Id
		return nil
	}else{
		return errors.New("DICT3 does not exist. Insert keys into it and later look them up")
	}
}

/*The listID function is used to return a list of unique IDs(key + relationship) from DICT3.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) ListIDs(input *JsonMessage, output *JsonListIDs) error {
	id := [][]string{}
	if _, err := os.Stat(serverconfig.PersistentStorageContainer.File); err == nil {
		inFile, err := os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_APPEND|os.O_CREATE,0660)
		checkError(err)
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		var DICT3exist DICT3format
		for scanner.Scan() {
			err := json.Unmarshal([]byte(scanner.Text()),&DICT3exist)
			checkError(err)
			tempid := []string{DICT3exist.Key,DICT3exist.Relationship}
			id = append(id,tempid)
		}
		output.Result = id
		output.Id = input.Id
		return nil
	}else{
		return errors.New("DICT3 does not exist. Insert values into it and later look them up")
	}
}

/*The shutdown function is used to shutdown the server process.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Shutdown(input *JsonMessage, output *NoOutput) error {
	close = true
	return nil
}



func main() {

	dict3 := new(Dict3)
	rpc.Register(dict3)
	
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "Enter config Json file path")
		log.Fatal(1)
	}
	
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		checkError(err)
	}
	
	//Unmarshalling the server config JSON file.
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()),&serverconfig)
		if err != nil {
			checkError(err)
		}
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(serverconfig.Port))
	checkError(err)
	
	//Listening for any active tcp connection at the specified port address.
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for !close {
		conn, _ := listener.Accept()
		jsonrpc.ServeConn(conn)
	}
}

/* This function is used to update the DICT3 file with any new insertions or updations of the triplet values.
* fileName: It refers to the file where all the changes goes and is mostly DICT3.
* key: The data that needs to be added to the file.*/
func saveJSON(fileName string, key interface{}) {
	outFile, err :=  os.OpenFile(fileName, os.O_RDWR|os.O_APPEND|os.O_CREATE,0660)
	checkError(err)
	encoder := json.NewEncoder(outFile)
	err = encoder.Encode(key)
	checkError(err)
	outFile.Close()
}

/* This function is used to check for any errors returned by the function.
* error: It refers to the error value that needs to be checked. */
func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
