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
	"time"
	"encoding/json"
	"strconv"
	"log"
	"math"
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
	DeleteTimeOut int `json:"deletetimeout"`
	Methods []string `json:"methods"`
}

/* The JSON message structure of the input passed to the function.
 * Method: The function to be called.
 * Params: The input json object that is passed as an argument to the function.
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
 * Error: The error that is returned by the function call. */
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
	Result [][]string `json:"result"`
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

/*The structure of the DICT3 file.
*Key: Refers to the key value of the parameter.
*Relationship: Refers to the relationship value of the parameter. Key and Relationship value together identifies an unique row of the DICT3 file.
*Value: Refers to the value of the argument passed to the function. It is a JSON object. */
type DICT3format struct{
	Key,Relationship string
	Value string
}

/*This structure is used when there is no need to display any output JSON message to the user.*/
type NoOutput struct {
	Error string
}

/*This structure is used to define the "key" components of the dictionary.*/
type datakey struct{
	key string
	relation string
}

/*This structure is used to define the "value" components of the dictionary.*/
type datavalue struct{
	content string
	size string
	created string
	modified string
	accessed string
	permission string
}

/*This structure is used to define all of the components of each server instance.*/
type Server struct {
	portno int
	successor int
	predecessor int
	fingertable map[int]int
	data map[datakey]datavalue
}

/*These are all the variables and flags that are used.*/
var serverconfig config
var count int
var count_of_server int
var use_ports int
var timediff time.Duration
var servermap map[int]Server
var ringmap map[int]int
var ringval int
var ringsize int
var dupver []int
var close map[int]bool


/*The lookUp function is used to return the value referred by an existing ID(key + relationship).
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) LookUp(input JsonMessage, output *JsonResultLookUp) error {
	if len(string(input.Params[0].(string))) == 0 && len(string(input.Params[1].(string))) == 0  {
			return errors.New("Key error - Both Key and Relationship attributes cannot be null in the input")
	}
	var targetport int
	key := string(input.Params[0].(string))
	relationship := string(input.Params[1].(string))
	datahash := DataHash(key,relationship)
	if (len(key) != 0 && len(relationship) != 0) {
		targetport = FindSuccessor(datahash,input.Portno)
		dupver = dupver[:0]
		storeddata := servermap[targetport].data
		for k,v:= range storeddata {
			if k.key == key && k.relation == relationship {
				servermap[targetport].data[k] = datavalue{v.content,v.size,v.created,v.modified,time.Now().Format("01/02/2006, 15:04:05"),v.permission}
				output.Result = append(output.Result,[]string{key,relationship,v.content})
				return nil
			}
		}
	}else if(len(key) != 0) {
		for i := 0; i < 16; i++ {
			hash := datahash + i;
			targetport = FindSuccessor(hash,input.Portno)
			dupver = dupver[:0]
			storedata := servermap[targetport].data
			for k,v := range storedata {
				if k.key == key {
					servermap[targetport].data[k] = datavalue{v.content,v.size,v.created,v.modified,time.Now().Format("01/02/2006, 15:04:05"),v.permission}
					temp := []string{key,k.relation,v.content}
					set := false
					if len(output.Result) > 0 {
						for i := range output.Result {
							if((output.Result[i][0] == temp[0]) && (output.Result[i][1] == temp[1])) {
									set = true
							}
						}
					}
					if (!set) {
						output.Result = append(output.Result,temp)
					}
				}
			}
		}
		return nil
	} else {
		for i := 0; i <= 112; i = i+16 {
			hash := datahash + i;
			targetport = FindSuccessor(hash,input.Portno)
			dupver = dupver[:0]
			storedata := servermap[targetport].data
			for k,v := range storedata {
				if k.relation == relationship {
					servermap[targetport].data[k] = datavalue{v.content,v.size,v.created,v.modified,time.Now().Format("01/02/2006, 15:04:05"),v.permission}
					temp := []string{k.key,relationship,v.content}
					set := false
					if len(output.Result) > 0 {
						for i := range output.Result {
							if((output.Result[i][0] == temp[0]) && (output.Result[i][1] == temp[1])) {
									set = true
							}
						}
					}
					if(!set) {
						output.Result = append(output.Result,temp)
					}
				}
			}
		}
		return nil
	}
	return errors.New("Key and/or Relationship not found in DICT3")
}

/*The insert function is used to enter an unique triplet in the DICT3 file.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Insert(input *JsonMessage, output *JsonResultInsert) error {
		if len(string(input.Params[0].(string))) == 0 || len(string(input.Params[1].(string))) == 0  {
			return errors.New("Key error - Key or Relationship attributes cannot be null in the input")
		}
		var targetport int
		var size int
		key := string(input.Params[0].(string))
		relation := string(input.Params[1].(string))
		datahash := DataHash(key,relation)
		targetport = FindSuccessor(datahash,input.Portno)
		dupver = dupver[:0]
		DICT3input := DICT3format{strings.TrimSpace(string(input.Params[0].(string))),strings.TrimSpace(string(input.Params[1].(string))),string(input.Params[2].(string))}
		storeddata := servermap[targetport].data
		size = len(DICT3input.Value)/1000
		if(size == 0){
			size = 1
		}
		for k,_ := range storeddata {
			if k.key == DICT3input.Key && k.relation == DICT3input.Relationship {
				return errors.New("Key error - Key and Relationship already present in DICT3. Use insertOrUpdate function to change values of existing key.")
			}
		}
		servermap[targetport].data[datakey{DICT3input.Key,DICT3input.Relationship}] = datavalue{DICT3input.Value,strconv.Itoa(size)+"KB",time.Now().Format("01/02/2006, 15:04:05"),"",time.Now().Format("01/02/2006, 15:04:05"),strings.TrimSpace(input.Params[3].(string))}
		output.Result = true
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
	var targetport int
	var size int
	key := string(input.Params[0].(string))
	relation := string(input.Params[1].(string))
	datahash := DataHash(key,relation)
	targetport = FindSuccessor(datahash,input.Portno)
	dupver = dupver[:0]
	DICT3input := DICT3format{strings.TrimSpace(string(input.Params[0].(string))),strings.TrimSpace(string(input.Params[1].(string))),string(input.Params[2].(string))}
	storeddata := servermap[targetport].data
	updated := false
	size = len(DICT3input.Value)/1000
	if(size == 0){
		size = 1
	}
	for k,_ := range storeddata {
		if k.key == DICT3input.Key && k.relation == DICT3input.Relationship {
			oldvalue := servermap[targetport].data[datakey{DICT3input.Key,DICT3input.Relationship}]
			if oldvalue.permission == "RW" {
				servermap[targetport].data[datakey{DICT3input.Key,DICT3input.Relationship}] = datavalue{DICT3input.Value,strconv.Itoa(size)+"KB",oldvalue.created,time.Now().Format("01/02/2006, 15:04:05"),time.Now().Format("01/02/2006, 15:04:00"),strings.TrimSpace(input.Params[3].(string))}
				updated = true
			} else {
					return errors.New("Permission Error - This value is Read only and cannot be updated.")
			}
		}
	}
	if updated == false {
		servermap[targetport].data[datakey{DICT3input.Key,DICT3input.Relationship}] = datavalue{DICT3input.Value,strconv.Itoa(size)+"KB",time.Now().Format("01/02/2006, 15:04:05"),"",time.Now().Format("01/02/2006, 15:04:05"),strings.TrimSpace(input.Params[3].(string))}
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
	var targetport int
	key := strings.TrimSpace(string(input.Params[0].(string)))
	relationship := strings.TrimSpace(string(input.Params[1].(string)))
	datahash := DataHash(key,relationship)
	targetport = FindSuccessor(datahash,input.Portno)
	dupver = dupver[:0]
	storeddata := servermap[targetport].data
	for k,v:= range storeddata {
		if k.key == key && k.relation == relationship {
			if v.permission == "RW" {
				delete(servermap[targetport].data,k)
				output.Error = " "
				return nil
			} else {
				return errors.New("Permission Error - This value is Read only and cannot be deleted.")
			}
		}
	}
	return errors.New("Key and/or Relationship not found in DICT3")
}

/*The listKey function is used to return a list of unique key values from DICT3.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) ListKeys(input *JsonMessage, output *JsonListKeys) error {
	keymap := make(map[datakey]struct{})
	key := []string{}
	for _,server := range servermap {
		storeddata := server.data
		for k,_ := range storeddata {
			keymap[k] = struct{}{}
		}
		for k := range(keymap){
			set := false
			if len(key) >= 1 {
				for i := range key {
					if (key[i] == k.key){
						set = true
					}
				}
			}
			if !set {
				key = append(key,k.key)
			}
		}
	}
	output.Result = key
	return nil
}

/*The listID function is used to return a list of unique IDs(key + relationship) from DICT3.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) ListIDs(input *JsonMessage, output *JsonListIDs) error {
	id := [][]string{}
	for _,server := range servermap {
		storeddata := server.data
		for k,_ := range storeddata {
			tempid := []string{k.key,k.relation}
			id = append(id,tempid)
		}
	}
	output.Result = id
	return nil
}

/*The purge function is used to remove stale entries from the dictionary which have not been accessed for specific time.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Purge(input *JsonMessage, output *NoOutput) error {
	for _,value := range servermap {
		for k,v := range value.data {
			datatime,_ := time.Parse("01/02/2006, 15:04:05",v.accessed)
			current := time.Now().UTC().Add(-4*60*time.Minute)
			diff := current.Sub(datatime)
			if  diff >= timediff {
				delete(value.data,k)
			}
		}
	}
	output.Error = " "
	return nil
}

/*The shutdown function is used to shutdown the server process.
*input: This input refers to the JsonMessage structure and has all the input arguments required by the function.
*output: It refers to the pointer structure that provides the output of the function.
*error: It contains the error value of the function if any error is generated.*/
func (d *Dict3) Shutdown(input *JsonMessage, output *NoOutput) error {
	succport := servermap[input.Portno].successor
	if (succport != input.Portno) {
		for id,value := range servermap[input.Portno].data {
			servermap[succport].data[id] = value
		}
		delete(servermap,input.Portno)
		for k,v := range ringmap {
			if v == input.Portno {
				delete(ringmap,k)
			}
		}
		StabilizeRing()
	}
	close[input.Portno] = false
	count++
	if count == count_of_server {
		fmt.Println("All the active servers have been closed.")
		fmt.Println("Saving all the data to the disk...")
		for _,value := range servermap {
			if(len(value.data) > 0){
				outFile, err :=  os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_RDWR|os.O_APPEND|os.O_CREATE,0660)
				checkError(err)
				for k,v := range value.data {
					if _,er := outFile.WriteString(k.key+"\t"+k.relation+"\t"+v.content+"\t"+v.size+"\t"+v.created+"\t"+v.modified+"\t"+v.accessed+"\t"+v.permission+"\n");er != nil{
						checkError(er)
						}
					}
				}
			}
		os.Exit(0)
	}
	return nil
}

/*The contains function is used to check whether a slice contains an integer or not.
*input: This input refers to the input slice list and the element to check whether it is present in the list.
*output: It refers to the boolean which indicates whether the element is present in the list or not.*/
func contains(s []int, e int) bool {
    for _, a := range s { if a == e { return true } }
    return false
}

/*This function is used to find the successor of the given key from the starting portno.
*input: The input refers to the key for which the successor from the given port no.
*output: The successor for the given key is returned.*/
func FindSuccessor(key int,portno int) int{
	var portlist []int
	succmap := make(map[int]int)
	port := portno
	set := false
	for !set {
		finger := servermap[port].fingertable
		used := false
		for k,v := range finger {
			if key >= k {
				diff := key - k
				succmap[diff] = v
				used = true
			}
		}
		if !used {
			port = servermap[port].successor
			continue
		}
		min := 200
		for k,_ := range succmap {
			if min > k{
				min = k
			}
		}
		if contains(portlist,succmap[min]) {
			set = true
			break
		}
		port = succmap[min]
		portlist = append(portlist,port)
	}
	min := 200
	for k,_ := range succmap {
		if min > k{
			min = k
		}
	}
	return succmap[min]
}

/*This function refers to converting the data to its corresponding hash value.
*input: The input refers to the key and relationship of the data.
*output: The output refers to the hash value that is generated from the given input.*/
func DataHash(key,rel string) int{
	nonce := []byte("875")
  sumkey := 0
  keybytes := []byte(key)
  relbytes := []byte(rel)
  for i:= 0; i < len(keybytes);i++ {
    sumkey += int(keybytes[i]) * int(nonce[(i%3)])
  }
  sumkey %= 128
  sumrel := 0
  for j := 0; j < len(relbytes); j++ {
    sumrel += int(relbytes[j]) * int(nonce[(j%3)])
  }
  sumrel %= 128
	hash := int((sumkey >> (8-4) << 4) + int(sumrel & 0x0f))
	return hash
}

/*The contains function that is used to stabilize a ring when a node enters or leaves the chord ring.
*The successor and predecessor for each of the chord ring nodes are calculated after they have been
fitted into the chord ring. The finger table for each of the nodes is also calculated.*/
func StabilizeRing() bool {
	for key,val := range ringmap {
		//To find the successor for each of the element in the ring
		temp := servermap[val]
		pos := key
		for true {
			if value,ok := ringmap[(pos+1)%128];ok{
				temp.successor = value
				break
			} else {
				pos++
				continue
			}
		}

		//To find the predecessor of the element in the ring
		pos = key
		for true {
			if (pos < 0) {
				pos = 127
			}
			if value,ok := ringmap[(pos-1)];ok {
				temp.predecessor = value
				break
			} else {
				pos--
				continue
			}
		}

		//To fill finger table for each of the ring elements
		for i:=0;i<7;i++ {
			pos = (key + int(math.Pow(2,float64(i))))%128
			fingertable_key := pos
			for true {
				if value,ok := ringmap[(pos)%128];ok {
					temp.fingertable[fingertable_key] = value
					break
				} else {
					pos++
					continue
				}
			}
		}
		servermap[val] = temp
	}
	return true
}

/*This contains the function that is used to create the hash for the input nodet to be inserted into the chord ring.
*input: The only input is the port no of the starting server.
*output: The computed hash of the incoming port.*/
func createhash(portno string) int {
  nonce := []byte("8757")
  inputport := []byte(portno)
  hash := 0
  for i := 0;i < len(inputport);i++ {
    hash += int(inputport[i]) * int(nonce[i])
  }
  hash %= ringsize
  return hash
}

func InitializeRing(){
  ringsize = 128
	ringval = 0
  ringmap = make(map[int]int)
	close = make(map[int]bool)
}

/*This contains a function that is used to add the server to the chord ring.
*input: This input refers to the number of servers to add to the ring.*/
func AddtoRing(in Server) {
	inHash := createhash(strconv.Itoa(in.portno))
  if _,ok := ringmap[inHash];ok {
    for i := ringval; i < ringsize; i++{
      if _,yes :=  ringmap[i]; yes {
        continue
      }
      ringmap[i] = in.portno
      ringval = i
      break
    }
  } else {
    ringmap[inHash]=in.portno
  }
}

/*This contains a function that is used to start the server with the corresponding Port number.
*input: This input refers to the port number of the server*/
func startserver(portno int) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(portno))
	checkError(err)
	//Listening for any active tcp connection at the specified port address.
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for close[portno] {
		conn, _ := listener.Accept()
		jsonrpc.ServeConn(conn)
	}
}

/*This contains a function that is used to add the server to the chord ring.
*input: This input refers to the number of servers to add to the ring.*/
func newServerInstance(n int) {
	for i := 1;i <= n;i++ {
		servermap[use_ports] = Server{use_ports,0,0,make(map[int]int),make(map[datakey]datavalue)}
		close[use_ports] = true
		go startserver(use_ports)
		AddtoRing(servermap[use_ports])
		count_of_server++
		use_ports++
	}
	if(len(ringmap) > 1){
		done := StabilizeRing()
		succport := servermap[use_ports-1].successor
		succdata := servermap[succport].data
		temp := servermap[succport]
		temp.data  = make(map[datakey]datavalue)
		servermap[succport] = temp
		for k,v := range succdata {
			datahash := DataHash(k.key,k.relation)
			targetport := FindSuccessor(datahash,use_ports-1)
			dupver = dupver[:0]
			servermap[targetport].data[datakey{k.key,k.relation}] = v
		}
		if done {
			fmt.Println("The ring has stabilized")
		}
	}
}

/*This is the main function that is used to get the input from the user.
*It is used to display the list of active servers and also to add a new server to the chord ring if needed.*/
func main(){
	var nodes int
	var choice int
	var port int
	servermap = make(map[int]Server)
	dict3 := new(Dict3)
	rpc.Register(dict3)
	count = 0
	count_of_server = 0
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
	use_ports = serverconfig.Port
	timediff = time.Duration(serverconfig.DeleteTimeOut) * time.Second
	InitializeRing()
	newServerInstance(1)
	fmt.Println("Enter the number of nodes to start the system")
	fmt.Scanf("%d",&nodes)
	newServerInstance(nodes)
	for true {
		fmt.Println();
		fmt.Println("1. Add a node to the system")
		fmt.Println("2. List all currently running servers")
		fmt.Println("3. Display the data present in the server.")
		fmt.Println("4. Display Position in Chord Ring, Successor Node, Predecessor Node and Finger Table for a server")
		fmt.Println("5. Exit")
		fmt.Println("Enter the choice:")
		fmt.Scanf("%d",&choice)
		switch choice {
				case 1: newServerInstance(1)
				case 2: fmt.Println();
								fmt.Println("The list of currently running servers with their Port Nos are as below-")
								for k,_ := range servermap {
									fmt.Println(k)
								}
				case 3: fmt.Println("Enter the port number of the server to check the data.")
								fmt.Scanf("%d",&port)
								for k,v := range servermap{
									if k == port {
										if len(v.data) > 0{
											fmt.Println();
											fmt.Println("The data present in the server is as below-");
											for key,val := range v.data {
												fmt.Print(key.key,"\t",key.relation)
												fmt.Print("\t",val.content,"\t",val.size,"\t",val.created,"\t",val.modified,"\t",val.accessed,"\t",val.permission)
												fmt.Println();
											}
										}else{
											fmt.Println("The server does not currently have any stored data.");
										}
									break
									}
								}
				case 4: fmt.Println("Enter the port number of the server.")
								fmt.Scanf("%d",&port)
								for k,v := range servermap{
									if k == port {
										for pos,port := range ringmap {
											if port == k {
												fmt.Println("Position in Chord Ring - ",pos)
												break
											}
										}
										fmt.Println("Successor Node - ",v.successor)
										fmt.Println("Predecessor Node- ",v.predecessor)
										fmt.Println("Finger Table - ")
										for key,value := range v.fingertable {
											fmt.Println(key,"\t",value)
										}
										break
									}
								}
				case 5:	var ch string
								fmt.Println("Do you want to save the data stored in the server? 'y' or 'n'")
								fmt.Scanf("%s",&ch)
								ch = strings.ToLower(ch)
								if(ch == "y"){
									for _,value := range servermap {
										if(len(value.data) > 0){
											outFile, err :=  os.OpenFile(serverconfig.PersistentStorageContainer.File, os.O_RDWR|os.O_APPEND|os.O_CREATE,0660)
											checkError(err)
											for k,v := range value.data {
												if _,er := outFile.WriteString(k.key+"\t"+k.relation+"\t"+v.content+"\t"+v.size+"\t"+v.created+"\t"+v.modified+"\t"+v.accessed+"\t"+v.permission+"\n");er != nil{
													checkError(er)
												}
											}
										}
									}
								}
								os.Exit(0)
				default: fmt.Println("The entered choice is invalid")
		}
	}
	var input string
	fmt.Scanln(&input)
}


/* This function is used to check for any errors returned by the function.
* error: It refers to the error value that needs to be checked.*/
func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
