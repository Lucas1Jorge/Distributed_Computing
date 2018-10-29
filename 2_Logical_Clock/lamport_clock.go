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
var ports []string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores dos outros processos
var ServerConn *net.UDPConn //conexão do meu servidor (onde recebo mensagens dos outros processos)
var times []int
var ids []int

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

    //Escreve na tela a msg recebida
    // fmt.Println(string(buf[0:n]), addr)
    _ = addr
    fmt.Println(string(buf[0:n]))

    if err != nil {
        fmt.Println("Error: ",err)
    } 
}

func doClientJob(otherProcess int, time int) {
    times[otherProcess - 1] = 1 + max(times[otherProcess - 1], time)

    msg := "Process " + ports[otherProcess - 1] + "\treceived at time " + strconv.Itoa(times[otherProcess - 1]) + "\tfrom " + ports[0]

    buf := []byte(msg)
    
    //otherServer
    _,err := CliConn[otherProcess-1].Write(buf)
    
    if err != nil {
        fmt.Println(msg, err)
    }
}

func initConnections() {
    nServers = len(os.Args) - 2
    for i := 0; i < nServers; i++ {
        ports = append(ports, os.Args[i+2])
    }
    /* Esse 2 tira o nome (no caso Process) e tira a primeira porta (que é a minha). As demais portas são dos outros processos*/

    // Outros códigos para deixar ok as conexões com os servidores dos outros processos
    /* Lets prepare a address at any address at port 10001*/   

    for i := 0; i < nServers; i++ {
        ServerAddr,err := net.ResolveUDPAddr("udp", ports[i])
        CheckError(err)
     
        // /* Now listen at selected port */
        ServerConn, err = net.ListenUDP("udp", ServerAddr)
        CheckError(err)

        LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        CheckError(err)

        Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
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

func main() {
    initConnections()
    
    // O fechamento de conexões devem ficar aqui, assim só fecha conexão quando a main morrer
    defer ServerConn.Close()

    for i := 0; i < nServers; i++ {
        defer CliConn[i].Close()
        times = append(times, 0)
    }

    // Todo Process fará a mesma coisa: ouvir msg e mandar infinitos i’s para os outros processos
    ch := make(chan string)

    // Read keyboard input
    go readInput(ch)

    for {
        // Server
    	go doServerJob()

		// When there is a request (from stdin). Do it!
		select {
			case x, valid := <-ch:
			
			if valid {
                x_int, _ := strconv.Atoi(x)
                if x_int == 1 {
                    times[0] += 1
                } else {
                    go doClientJob(x_int, times[0])
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