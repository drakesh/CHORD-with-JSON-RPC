package main

import (
	"net/rpc/jsonrpc"
	"fmt"
	"log"
	"os"
	"bufio"
	"encoding/json"
	"strconv"
)

/* The JSON message structure of the input passed to the program.
 * Method: Remote function to be called.
 * Params: The input json object that is passed to the function.
 * Id: The unique id refers to the transaction between the client and the server.*/
type JsonMessage struct{
	Method string 	`json:"method"`
	Params []interface{} `json:"params"`
	Portno int `json:"port"`
}

/* The json result structure that will be displayed to the user after the completion
 * of the insert function.
 * Result: Indicates the successful or unsuccessful completion of the function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonResultInsert struct{
	Result bool `json:"result"`
	Error string `json:"error"`
}

/* The JSON result structure that will be displayed to the user after the completion
 * of the lookup function.
 * Result: Indicates the return of the params identified by the key and the relationship.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonResultLookUp struct{
	Result []interface{} `json:"result"`
	Error string `json:"error"`
}

/* The json result structure that will be displayed to the user after the completion
 * of the listKey function.
 * Result: Identifies the list of unique keys returned by the function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonListKeys struct{
	Result []string `json:"result"`
	Error string `json:"error"`
}

/* The JSON result structure that will be displayed to the user after the completion
 * of the listID function.
 * Result: Identifies the list of unique key and relationship combination returned by the listID function.
 * Id: The unique id refers to the transaction between the client and server.
 * Error: The error that is returned by the remote function call. */
type JsonListIDs struct{
	Result [][]string `json:"result"`
	Error string `json:"error"`
}

/*This structure is used when there is no need to display any output JSON message to the user. */
type NoOutput struct {
	Error string
}

/*The refers to the input JSON message structure that represents the configuration details of the client.
 * Server ID: Refers to the remote server ID.
 * Protocol: Refers to the protocol used to contact the remote server and is mostly TCP.
 * IPAddress: Refers to the IP address of the remote server.
 * Port: Refers to the port number being used to start the communication process.
 * Methods: Contains the list of all the functions that can be called at the remote server. */
type config struct{
	ServerId string `json:"serverID"`
	Protocol string `json:"protocol"`
	IpAddress string `json:"ipAddress"`
	Port int `json:"port"`
	Methods []string `json:"methods"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "Enter config Json file path")
		log.Fatal(1)
	}
	inFile, err := os.Open(os.Args[1])
	if err != nil {
			log.Fatal("Opeining the Json Config file :", err)
	}
	scanner := bufio.NewScanner(inFile)

	//Unmarshalling the input config fille.
	var clientconfig config
	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()),&clientconfig)
		if err != nil {
			log.Fatal("Error Unsmrashalling the Json config file:", err)
		}
	}
	service := clientconfig.IpAddress +":" + strconv.Itoa(clientconfig.Port)

	//Refers to the establishment of the tcp connection between the client and the server.
	client, err := jsonrpc.Dial(clientconfig.Protocol, service)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	if client == nil{
		log.Fatal("dialing:", err)
	}
	var JsonInput JsonMessage
	JsonInput.Portno = clientconfig.Port
	scanner = bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
	  text := scanner.Text()

		//Unmarshalling the input JSON message.
		err = json.Unmarshal([]byte(text),&JsonInput)
		if err != nil {
			log.Fatal("Json Error:", err)
		}

		//Calling the appropriate function that is referred in the input JSON message.
		//The result is displayed in the console depending on the function that is being called.
		switch{
		case JsonInput.Method == "lookup":
			resultlookup := new(JsonResultLookUp)
			lookupcall := client.Go("Dict3.LookUp",JsonInput,resultlookup,nil)
			replycall := <-lookupcall.Done
			if replycall.Error == nil {
				resultlookup.Error = "null"
			}else{
				resultlookup.Error = replycall.Error.Error()
			}
			JsonOutput, err := json.Marshal(resultlookup)
			if err != nil {
				log.Fatal("Marshaling the result to display:", err)
			}
			fmt.Printf("%s\n",JsonOutput)

		case JsonInput.Method == "insert":
			resultinsert := new(JsonResultInsert)
			insertcall := client.Go("Dict3.Insert",JsonInput,resultinsert,nil)
			replycall := <-insertcall.Done
			if replycall.Error == nil {
				resultinsert.Error = "null"
			}else{
				resultinsert.Result = false
				resultinsert.Error = replycall.Error.Error()
			}
			JsonOutput, err := json.Marshal(resultinsert)
			if err != nil {
				log.Fatal("Marshaling the result to display:", err)
			}
			fmt.Printf("%s\n",JsonOutput)
		case JsonInput.Method == "insertOrUpdate":
			nooutput := new(NoOutput)
			insertorupdatecall := client.Go("Dict3.InsertOrUpdate",JsonInput,nooutput,nil)
			replycall := <-insertorupdatecall.Done
			if replycall.Error != nil {
				nooutput.Error = replycall.Error.Error()
				JsonOutput, err := json.Marshal(nooutput)
				if err != nil {
					log.Fatal("Marshaling the result to display:", err)
				}
				fmt.Printf("%s\n",JsonOutput)
			}
		case JsonInput.Method == "delete":
			nooutput := new(NoOutput)
			deletecall := client.Go("Dict3.Delete",JsonInput,nooutput,nil)
			replycall := <-deletecall.Done
			if replycall.Error != nil {
				nooutput.Error = replycall.Error.Error()
				JsonOutput, err := json.Marshal(nooutput)
				if err != nil {
					log.Fatal("Marshaling the result to display:", err)
				}
				fmt.Printf("%s\n",JsonOutput)
			}
		case JsonInput.Method == "listKeys":
			resultlistkeys := new(JsonListKeys)
			listkeyscall := client.Go("Dict3.ListKeys",JsonInput,resultlistkeys,nil)
			replycall := <-listkeyscall.Done
			if replycall.Error == nil {
				resultlistkeys.Error = "null"
			}else{
				resultlistkeys.Error = replycall.Error.Error()
			}
			JsonOutput, err := json.Marshal(resultlistkeys)
			if err != nil {
				log.Fatal("Marshaling the result to display:", err)
			}
			fmt.Printf("%s\n",JsonOutput)
		case JsonInput.Method == "listIDs":
			resultlistIDs := new(JsonListIDs)
			listIDscall := client.Go("Dict3.ListIDs",JsonInput,resultlistIDs,nil)
			replycall := <-listIDscall.Done
			if replycall.Error == nil {
				resultlistIDs.Error = "null"
			}else{
				resultlistIDs.Error = replycall.Error.Error()
			}
			JsonOutput, err := json.Marshal(resultlistIDs)
			if err != nil {
				log.Fatal("Marshaling the result to display:", err)
			}
			fmt.Printf("%s\n",JsonOutput)
		case JsonInput.Method == "purge":
			nooutput := new(NoOutput)
			purgeCall := client.Go("Dict3.Purge",JsonInput,nooutput,nil)
			purgeresult := <-purgeCall.Done
			if purgeresult.Error == nil {
				fmt.Println("The CHORD ring was successfully purged of any stale entries.")
			}else{
				fmt.Println("There was an error purging the CHORD ring - ",purgeresult.Error)
			}
		case JsonInput.Method == "shutdown":
			nooutput := new(NoOutput)
			shutdowncall := client.Go("Dict3.Shutdown",JsonInput,nooutput,nil)
			_ = <-shutdowncall.Done
		default:
			fmt.Println("The input method does not exist. Check the config file to see which method exists at the server")
		}
	}
}
