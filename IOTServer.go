package main
/****
*
* In memory transaction server that calculates tag position; Utilizes Kalman filters to assist with location of the tag.
* - Expects the following tables:
*
*   Table 1: Installations
*   +-------+--------+--------------+--------+-----------+------------+-----------+
*   | id    | Memint | Access table | Name   | Location  | NodeChart  | Tag chart |
*   ------------------------------------------------------------------------------
*   
*   Table 2: Tags
*   +-------+--------+-----------+---------+
*   |id     |memint  |Tag chart  | MAC     |
*   ---------------------------------------
*
*   Table 3: Nodes
*   +-------+--------+-----------+---------+-----------+
*   |id     |loc     |RSSISLOT   | MAC     | Last Call |
*   ---------------------------------------------------
*/
import (
    "crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"crypto/x509"
	"io/ioutil"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)
//This struct represents 1 ibeacon tag
type Element struct {
    MAC [17]byte //This is the bluetooth MAC of our tag
    RSSI []int //This is the RSSI as read by the nodes associated with an installation
    DelStack []int //This is for the remove reading function to handle so as to prevent unwanted deletes from muotiple threads
    installation int //This is the installation it belongs to
    taglock int //This is the lock to ensure that the location is only updated as quickly as possible
}
alltags [][]Element //Indexes stacks via their installation thus minimizing search time
var db;

const ConnectionType (
    AlertingNode int iota
    ReportingNode
    )

func initInMemoryTransactions(){
    //Start by reading database
	db, err := sql.Open("mysql",
		":password@tcp(127.0.0.1:3306)/hello")
	if err != nil {
		log.Fatal(err)
	}

}
func getLocation(SenderMAC [17]byte) (error err, int loc,int slot) {
	rows, err = db.Query("SELECT loc, RSSISLOT FROM NODES WHERE MAC=?",string(SenderMAC));
	defer rows.close()
	if (err != nil) {
		return (err, -1,-1)
	}
	var location int
	var rssi int
	for rows.Next() {
		err := rows.Scan(&location, &rssi)
		if err != nil {
			return (err,-1,-1)
		}
		return (nil, location, rssi)
	}
	return (error.New("Cannot locate"),-1,-1)
}
}
func determineIfTrackable(MAC [17]byte) bool {
    //determine if trackable
    return true
}

//Optimization function search from back to compare two macs
func compareMAC(MAC1 [17]byte, MAC2 [17]byte) bool {
    for i := 16 ; i >=0 ; i-- {
        if MAC1[i]!=MAC[i] {
            return false
        }
    }
    return true
}

func prepareDelete(int location ,int element, int rssislot){
    //Claim the value through DelStack belongs to this request
    alltags[location][element].DelStack[rssislot]++
    //Job done lets sleep
    timer := time.NewTimer(time.Second * 120)
    <- timer.C
    //release claim on the rssi after 2 min
    alltags[location][element].DelStack[rssislot]--
    if alltags[location][element].DelStack[rssislot] < 1 {
        //Check if this value belongs to the original request that spawned this thread. If it does delete it
        alltags[location][element].RSSI[rssislot] = 0
    }
}

//This function handles incoming BLE messages
//
func handleIncomingMessage(s string) error {
    SenderMAC := [17]byte(s[:17])
    TargetMAC := [17]byte(s[17:34])
    RSSI      := strconv.Atoi(s[34:39])
    err,location,rssislot := getLocation(SenderMAC)
    if err != nil {
    	return err
    }
    if determineIfTrackable(TargetMAC) {
        for i:=0; i < len(alltags); i++{
            // Go through the linked list 
            if compareMAC(alltags[location][i].MAC, TargetMAC) { 
                i := len(alltags)
                alltags[location][i].RSSI[rssislot] := TargetMAC
                go prepareDelete(location,i,rssislot)
            }
        }    
    }
    return nil
}
func determineConnectionType(connection string){
    
}
//The main IOT server
func IOTServer(){
    //Load the CA file
    ca_b, _ := ioutil.ReadFile("ca.pem")
	ca, _ := x509.ParseCertificate(ca_b)
	priv_b, _ := ioutil.ReadFile("ca.key")
	priv, _ := x509.ParsePKCS1PrivateKey(priv_b)

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	cert := tls.Certificate{
		Certificate: [][]byte{ ca_b },
		PrivateKey: priv,
	}

    //Verify the client
	config := tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs: pool,
	}
	
	config.Rand = rand.Reader
	service := "0.0.0.0:7176"
	listener, err := tls.Listen("tcp", service, &config)
	
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleClient(conn)
	}
}

//This handles individual IOT connections
func handleClient(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	
	for {
		log.Print("server: conn: waiting")
		n, err := conn.Read(buf)
		if err != nil {
			if err != nil {
				log.Printf("IOTServer.go: conn: read: %s", err)
			}
			break
		}

		tlscon, ok := conn.(*tls.Conn)
		if ok {
			state := tlscon.ConnectionState()
			sub := state.PeerCertificates[0].Subject
			log.Println(sub)
		}

		log.Printf("server: conn: echo %q\n", string(buf[:n]))
		n, err = conn.Write(buf[:n])

		n, err = conn.Write(buf[:n])
		log.Printf("server: conn: wrote %d bytes", n)

		if err != nil {
			log.Printf("server: write: %s", err)
			break
		}
	}
	log.Println("server: conn: closed")
}

func main(){
    //Consists of TCP server for streaming location and TLS server for IOT
    go MapviewServer()
    go IOTServer()
}



