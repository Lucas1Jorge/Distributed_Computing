// Chang-Roberts election algorithm
// by..: lucas1jorge

package main
import (
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

    msg := string(buf[0:n])
    //Escreve na tela a msg recebida
    fmt.Println("P" + myPort[1:] + ": Received", msg)

    if err != nil {
        fmt.Println("Error: ",err)
    } 
    
    if msg[:1] == "s" {
        cand, _ := strconv.Atoi(msg[1:])

        if cand > id {
            candidate(cand)
        } else if cand < id {
            candidate(id)
        } else {
            elect(id)
        }
    } else if msg[:1] == "f" {
        elec, _ := strconv.Atoi(msg[1:])

        if elec > id {
            elect(elec)
        } else if elec < id {
            elect(id)
        } else {
            fmt.Println("Found a Leader: ", elec)
        }
    }
}

func doClientJob(otherProcess int, msg string) { //Envia uma mensagem "S" ou "F"
    buf := []byte(msg)
    
    //otherServer
    _,err := CliConn[otherProcess].Write(buf)
    
    if err != nil {
        fmt.Println(msg, err)
    }

    fmt.Println("P" + myPort[1:] + ": Sending " + msg + " to " + nextPort[1:])
}

func initConnections() {
    id, _ = strconv.Atoi(os.Args[2])
    myPort = os.Args[2 + id]
    nextPort = os.Args[3 + id]
    nServers = len(os.Args) - 3
    /* Esse 2 tira o nome (no caso Process) e tira a primeira porta (que é a minha). As demais portas são dos outros processos*/

    /* Lets prepare a address at any address at port 10001*/   
    ServerAddr, err := net.ResolveUDPAddr("udp", myPort)
    CheckError(err)

    /* Now listen at selected port */
    ServerConn, err = net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    // Outros códigos para deixar ok as conexões com os servidores dos outros processos
 
    for i := 0; i < nServers; i++ {
        if (i == id -1) {
            continue
        }

        LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        CheckError(err)

        ConnAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1" + os.Args[i+3])
        CheckError(err)

        Conn, err := net.DialUDP("udp", LocalAddr, ConnAddr)
        CheckError(err)

        CliConn = append(CliConn, Conn)
    }
}

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func candidate(cand int) {
    msg := "s" + strconv.Itoa(cand)

    np, _ := strconv.Atoi(nextPort)
    doClientJob(np, msg)
}

func elect(cand int) {
    msg := "f" + strconv.Itoa(cand)

    np, _ := strconv.Atoi(nextPort)
    doClientJob(np, msg)
}

func main() {
    // fmt.Println("OK 1")
    initConnections()
    
    // O fechamento de conexões devem ficar aqui, assim só fecha conexão quando a main morrer
    defer ServerConn.Close()

    for i := 0; i < nServers - 1; i++ {
        defer CliConn[i].Close()
    }

    // Todo Process fará a mesma coisa: ouvir msg e mandar infinitos i’s para os outros processos
    ch := make(chan string)

    // Read keyboard input
    go readInput(ch)

    for {
        go doServerJob()
		// When there is a request (from stdin). Do it!
		select {
			case x, valid := <-ch:
			
			if valid {
				fmt.Printf("Read from keyboard: %s \n", x)

                if x == "start" { // start election
                    candidate(id)
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
