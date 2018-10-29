package main
import (
    "encoding/json"
    "strings"
    "fmt"
    "net"
    "os"
    "strconv"
    "time"
    "bufio"
)

//Variáveis globais interessantes para o processo
var err string
var id int
var myPort string //porta do meu servidor
var nextPort string //porta do servidor seguinte
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores dos outros processos
var ServerConn *net.UDPConn //conexão do meu servidor (onde recebo mensagens dos outros processos)

var latest []int
var wait []bool
var num []int
var WF []int
var engager []int

type message struct {
    QR string
    I int
    M int
    J int
    K int
}

func max(x int, y int) int {
    if x > y {
        return x;
    } else {
        return y;
    }
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
    n,addr,err := ServerConn.ReadFromUDP(buf)
    _ = addr
    
    var msg message
    err = json.Unmarshal(buf[:n], &msg)
    CheckError(err)

    //Escreve na tela a msg recebida
    fmt.Println("P" + myPort[1:] + ": Received", string(buf[:n]))

    if msg.QR == "Q" { // Query
        if msg.M > latest[id-1] { // engaging query
            latest[id-1] = msg.M
            engager[id-1] = msg.J
            wait[id-1] = true
            num[id-1] = len(WF)
            for j := 0; j < len(WF); j++ {
                send("Q", msg.I, msg.M, msg.K, WF[j])
            }
        } else if wait[id-1] && msg.M == latest[id-1]{ // not engaging query
            send("R", msg.I, msg.M, msg.K, msg.J)
        }
    } else if msg.QR == "R" { // Reply
        if wait[id-1] && msg.M == latest[id-1] {
            num[id-1]--
            if num[id-1] == 0 {
                if msg.I == msg.K {
                    fmt.Println("P", id, "is Deadlocked !!")
                } else {
                    send("R", msg.I, msg.M, msg.K, engager[id-1])
                }
            }
        }
    }
}

func doClientJob(otherProcess int, msg message) {
    my_msg := message{msg.QR, msg.I, msg.M, id, otherProcess}

    jsonRequest, err := json.Marshal(my_msg)
    CheckError(err)

    _, err = CliConn[otherProcess-1].Write(jsonRequest)
    CheckError(err)

    fmt.Println("P" + myPort[1:] + ": Sending", string(jsonRequest), "to", otherProcess)
}

func initConnections() {
    id, _ = strconv.Atoi(os.Args[2])
    myPort = os.Args[2 + id]
    nServers = len(os.Args) - 3
    if (id == nServers) {
        nextPort = os.Args[3]
    } else {
        nextPort = os.Args[3 + id]
    }

    /* Lets prepare a address at any address at port 10001*/   
    ServerAddr, err := net.ResolveUDPAddr("udp", myPort)
    CheckError(err)

    /* Now listen at selected port */
    ServerConn, err = net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    // Outros códigos para deixar ok as conexões com os servidores dos outros processos
 
    LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    CheckError(err)

    for i := 0; i < nServers; i++ {
        // if (i == id -1) {
        //     continue
        // }

        ConnAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + os.Args[i+3])
        CheckError(err)

        Conn, err := net.DialUDP("udp", LocalAddr, ConnAddr)
        CheckError(err)

        CliConn = append(CliConn, Conn)
    }

    // Shared Resource Connection
    SharedResourceAdd, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
    CheckError(err)

    SharedResourceConn, err := net.DialUDP("udp", LocalAddr, SharedResourceAdd)
    CheckError(err)

    CliConn = append(CliConn, SharedResourceConn)
}

func readInput(ch chan string) {
    // Non-blocking async routine to listen for terminal input
    reader := bufio.NewReader(os.Stdin)
    for {
        text, _, _ := reader.ReadLine()
        ch <- string(text)
    }
}

func send(qr string, i int, m int, j int, k int) {
    doClientJob(k, message{qr, i, m, j, k})
}

func start() {
    latest[id-1]++
    m := latest[id-1]
    wait[id-1] = true
    num[id-1] = len(WF)

    for j := 0; j < len(WF); j++ {
        fmt.Println("Starting Q:", id, m, id, WF[j])
        send("Q", id, m, id, WF[j])
    }
}

func main() {
    initConnections()
    
    // close Conns when main dies
    defer ServerConn.Close()

    for i := 0; i < nServers; i++ {
        defer CliConn[i].Close()
    }

    ch := make(chan string)

    // Read keyboard input
    go readInput(ch)

    latest = make([]int, nServers)
    wait = make([]bool, nServers)
    num = make([]int, nServers)
    // WF = make([]int, nServers)
    engager = make([]int, nServers)

    for {
        go doServerJob()
        
        // When there is a request (from stdin). Do it!
        select {
            case x, valid := <-ch:
            
            if valid {
                fmt.Printf("Read from keyboard: %s \n", x)

                if x == "start" {
                    start()
                } else if len(x) >= 7 && x[:7] == "Process" {
                    wf := strings.Split(x, " ")
                    // fmt.Println(wf)
                    for j := 2; j < len(wf); j++ {
                        dependence, _ := strconv.Atoi(wf[j][1:])
                        WF = append(WF, dependence)
                    }
                    // fmt.Println(WF)
                } else {
                    fmt.Println("Unexpected input")
                }
            } else {
                fmt.Println("Channel closed!")
            }
            default:
                // Do nothing in the non-blocking approach.
                time.Sleep(time.Second * 1)
            }

            // Wait a while
            time.Sleep(time.Second * 1)
        }
}
