package main
import (
    "fmt"
    "net"
    "os"
    "strconv"
    "time"
)

//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores dos outros processos
var ServerConn *net.UDPConn //conexão do meu servidor (onde recebo mensagens dos outros processos)

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
    n,addr,err := ServerConn.ReadFromUDP(buf)

    //Escreve na tela a msg recebida
    fmt.Println("Received ",string(buf[0:n]), " from ",addr)

    if err != nil {
        fmt.Println("Error: ",err)
    } 
}

func doClientJob(otherProcess int, i int) {
    //Envia uma mensagem (com valor i) para o servidor do processo
    msg := strconv.Itoa(i)

    buf := []byte(msg)
    
    //otherServer
    _,err := CliConn[otherProcess].Write(buf)
    
    if err != nil {
        fmt.Println(msg, err)
    }
}

func initConnections() {
    myPort = os.Args[1]
    nServers = len(os.Args) - 2
    /* Esse 2 tira o nome (no caso Process) e tira a primeira porta (que é a minha). As demais portas são dos outros processos*/

    // Outros códigos para deixar ok as conexões com os servidores dos outros processos
    /* Lets prepare a address at any address at port 10001*/   
    ServerAddr,err := net.ResolveUDPAddr("udp",":10001")
    CheckError(err)
 
    // /* Now listen at selected port */
    ServerConn, err = net.ListenUDP("udp", ServerAddr)
    CheckError(err)

    LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)
 
    for i := 0; i < nServers; i++ {
        Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
        CheckError(err)

        CliConn = append(CliConn, Conn)
    }
}

func main() {
    // fmt.Println("OK 1")
    initConnections()
    
    // O fechamento de conexões devem ficar aqui, assim só fecha conexão quando a main morrer
    defer ServerConn.Close()

    for i := 0; i < nServers; i++ {
        defer CliConn[i].Close()
    }

    // Todo Process fará a mesma coisa: ouvir msg e mandar infinitos i’s para os outros processos
    i := 0
    for {
        // Server
        go doServerJob()
        
        // Client
        for j := 0; j < nServers; j++ {
            go doClientJob(j, i)
        }
        
        // Wait a while
        time.Sleep(time.Second * 1)
        i++
    }
}