1. Starting the client and server process:
The client can be started by executing the ‘client’ shell script. The shell script takes care of adding the ‘clientconfig.json’ configuration file as a command line argument.
./client
The server can also be started in a similar way by executing the script. The script includes the command line argument JSON configuration file. It is as shown below -
./server
Now after starting both the client and the server process, you can give the input JSON message to the client.
When the server is started, it will ask for the number of nodes to start the Chord ring with. After the user has entered the number, the Chord ring will stabilize with the initial number of nodes. The server then displays the below options in the terminal –
1. Add a node to the system
2. List all currently running servers
3. Display the data present in the server
4. Display Position in Chord Ring, Successor Node, Predecessor Node and Finger Table
5. Exit
2. The server should at least be up and running before giving any inputs to the client. The input to the client is a JSON message and can be given in 2 ways  as below –
a. JSON message in the standard input:
You can type all the JSON message entirely in the command line and press enter or Ctrl+D to execute the input. It is as shown below –
   ./client
   {"method":"insertOrUpdate","params":["keyD","relA",{"a":1}],"id":5}
b. Using redirectional operator:
You can also pass in a file containing the JSON input messages using the redirectional operator ‘<’ as shown below –
   ./client < input.txt
   where, input.txt is the file containing the input JSON message.
3. After all the required operations have been performed, you can shut down all of the servers by either passing the ‘shutdown’ JSON message from the client on each of the server in the Chord ring or by going over to the server side and using option 5 to exit the Chord ring. The first option will by default persist the data to the disk and the second option will ask the user whether the data needs to be persisted on the disk or not.
The server and the client program has been tested on both Linux and Windows machine
