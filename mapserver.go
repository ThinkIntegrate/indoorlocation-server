package main
func handleMapRequest(client net.Conn) {
    b := bufio.NewReader(client)
    for {
        line, err := b.ReadBytes('\n')
        if err != nil { // EOF, or worse
            break
        }
        if(line){
            
        }
    }
}
func MapviewServer(){
    ln, _ := net.Listen("tcp", ":9123")
    conn, _ := ln.Accept()
    for {
        go handleMapRequest(<-conn)
    }
}