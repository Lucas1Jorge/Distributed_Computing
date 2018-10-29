package main
import (
    "encoding/json"
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
var my_time int
var my_T int
var n_replies int
var state string
var myPort string //porta do meu servidor
var nextPort string //porta do servidor seguinte
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores dos outros processos
var ServerConn *net.UDPConn //conexão do meu servidor (onde recebo mensagens dos outros processos)
var my_queue []message

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
    my_time = 1 + max(my_time, msg.T)

    if state == "WANTED" && msg.P == id {
        n_replies++
    } else if state == "RELEASED" ||
    (state == "WANTED" && msg.T < my_T) ||
    (state == "WANTED" && msg.T == my_T && msg.P < id){
        reply(msg.P)
    } else {
        msg.T = my_time
        enQueue(msg)
    }

    CheckError()
}

func doClientJob(otherProcess int, msg string) {
    p, _ := strconv.Atoi(msg)
    my_msg := message{my_time, p, "CS sugou"}

    jsonRequest, err := json.Marshal(my_msg)
    CheckError(err)

    // buf := []byte(msg)

    _, err = CliConn[otherProcess].Write(jsonRequest)
    
    if err != nil {
        fmt.Println(msg, err)
    }

    fmt.Println("P" + myPort[1:] + ": Sending", string(jsonRequest), "to", otherProcess + 1)
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

func request() {
    for i := 0; i < nServers; i++ {
        if i == id - 1 {
            continue
        }
        msg := strconv.Itoa(id)
        doClientJob(i, msg)
    }
}

func reply(P int) {
    msg := strconv.Itoa(P)

    doClientJob(P-1, msg)
}

func use_CS() {
    msg := strconv.Itoa(id)

    doClientJob(nServers, msg)

    time.Sleep(time.Second * 5)
}

func enQueue(msg message) {
    my_queue = append(my_queue, msg)
}

func release() {
    for _, v := range my_queue {
        msg := strconv.Itoa(v.P)

        doClientJob(v.P - 1, msg)
    }

    my_queue = my_queue[:0]
}

func main() {
    initConnections()
    
    // O fechamento de conexões devem ficar aqui, assim só fecha conexão quando a main morrer
    defer ServerConn.Close()

    for i := 0; i < nServers; i++ {
        defer CliConn[i].Close()
    }

    // Todo Process fará a mesma coisa: ouvir msg e mandar infinitos i’s para os outros processos
    ch := make(chan string)

    // Read keyboard input
    go readInput(ch)

    state = "RELEASED"
    n_replies = 0

    for {
        go doServerJob()

        if n_replies == nServers - 1 {
            fmt.Println("Entered CS")

            use_CS()
            n_replies = 0
            release()
            state = "RELEASED"

            fmt.Println("Left CS")
        }
        
        // When there is a request (from stdin). Do it!
        select {
            case x, valid := <-ch:
            
            if valid {
                fmt.Printf("Read from keyboard: %s \n", x)

                if state == "WANTED" || state == "HELD" {
                    fmt.Println(x, "ignored")
                } else { // state == "RELEASED"
                    if x == strconv.Itoa(id) { // start election
                        my_time++
                    } else if x == "x" {
                        state = "WANTED"
                        n_replies = 0
                        time.Sleep(time.Second * 5)
                        request()
                        my_T = my_time
                    } else {
                        fmt.Println("Unexpected input")
                    }
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

// test := make([]int, 2)
// test[0] = 1
// test[1] = 2
// fmt.Println(test)
