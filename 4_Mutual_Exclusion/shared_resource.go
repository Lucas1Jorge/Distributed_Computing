package main
import (
    "encoding/json"
    "fmt"
    "net"
    "os"
    // "strconv"
    "time"
    // "bufio"
)

//Variáveis globais interessantes para o processo
var err string
var nServers int //qtde de outros processo
var Connection *net.UDPConn //conexão do meu servidor (onde recebo mensagens dos outros processos)

type message struct {
    T int
    P int
    Text string
}

func CheckError(err error) {
    if err != nil {
        fmt.Println("Erro: ", err)
        os.Exit(0)
    }
}

func PrintError(err error) {
    if err != nil {
        fmt.Println("Erro: ", err)
    }
}

func doServerJob() {
	buf := make([]byte, 1024)   
 
	//Ler (uma vez somente) da conexão UDP a mensagem
    n,addr,err := Connection.ReadFromUDP(buf)
    _ = addr
    
    var msg message
    err = json.Unmarshal(buf[:n], &msg)
    CheckError(err)

    //Escreve na tela a msg recebida
    fmt.Println("Shared Resuorce received", string(buf[:n]))

    if err != nil {
        fmt.Println("Error: ",err)
    }
}

func initConnections() {
    Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)

	Connection, err = net.ListenUDP("udp", Address)
	CheckError(err)
}

func main() {
    initConnections()
    
    // O fechamento de conexões devem ficar aqui, assim só fecha conexão quando a main morrer
    defer Connection.Close()	

    for {
        go doServerJob()

        time.Sleep(time.Second * 1)
    }
}