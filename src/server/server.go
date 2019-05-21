package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

var clients []net.Conn
var matchs []Match
var playerNames = ""
var waitingRoom []net.Conn
var maxPlayersPerMatch = 2
var match Match
var guess string
var tip string

type Match struct {
	matchID int
	// playersCon     []net.Conn
	players        []Player
	idMasterPlayer int
	guess          string //Defined by the Master
	tip            string //Defined by the Master
	idTurnPlayer   int
	winner         *Player

	// idTurnPlayer   []string
	// nextTurnPlayer int
}

type Player struct {
	con   net.Conn
	name  string
	score int
}

func main() {
	var (
		host   = "127.0.0.1"
		port   = "8000"
		remote = host + ":" + port
		data   = make([]byte, 1024)
	)
	// timer := time.NewTimer(2 * time.Second)

	fmt.Println("Inicializando servidor... (Ctrl-C para encerrar)")
	match := Match{}

	lis, err := net.Listen("tcp", remote) //Começa a escutar no endereço 127.0.0.1:8000
	defer lis.Close()

	if err != nil {
		fmt.Printf("Erro no endereço: %s, Err: %s\n", remote, err)
		os.Exit(-1)
	}

	for {
		var res string
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Erro ao conectar o cliente : ", err.Error())
			os.Exit(0)
		}
		clients = append(clients, conn)
		waitingRoom = append(waitingRoom, conn)

		go func(con net.Conn) {

			// con.Write([]byte(msg))

			length, err := con.Read(data)
			name := string(data[:length])
			fmt.Println("Nova conexão : ", con.RemoteAddr())
			fmt.Println("O Jogador : ", name, "entrou no servidor!")
			if err != nil {
				fmt.Printf("O cliente %v saiu.\n", con.RemoteAddr())
				con.Close()
				disconnect(con, con.RemoteAddr().String())
				return
			}

			if playerNames == "" {
				playerNames = name
			} else {
				playerNames = playerNames + "," + name
			}
			comeStr := name + " entrou na sala de espera!\n"

			min := 0
			max := 2

			player := Player{}
			player.con = con
			player.name = name
			player.score = 0
			match.players = append(match.players, player)
			match.idMasterPlayer = rand.Intn(max-min) + min

			notifyAllButYou(con, comeStr, waitingRoom)

			if len(waitingRoom) < maxPlayersPerMatch {
				// notifyPlayer("serverMessage:"+"Atualmente há: "+strconv.Itoa(len(clients))+" jogadores conectados no servidor.\n", con)
				// <-timer.C
				notifyPlayer("serverMessage:"+strconv.Itoa(len(waitingRoom))+" jogadores na sala de espera.\n", con)

			}

			if len(waitingRoom) == maxPlayersPerMatch {

				notifyAll("initializeMatch:1:"+strconv.Itoa(len(waitingRoom))+":"+playerNames+":torres\n", waitingRoom)
				// <-timer.C
				// notifyAll("initializeMatch:1", waitingRoom)
				notifyPlayer("setMaster\n", waitingRoom[1])
				notifyAllButYou(waitingRoom[1], "serverMessage:O mestre está definindo o personagem a ser advinhado, aguarde sua vez.\n", waitingRoom)

				// 	caracter, err := waitingRoom[0].Read(data)
				// 	if err != nil {
				// 		fmt.Printf("Error")
				// 		con.Close()
				// 		disconnect(con, name)
				// 		return
				// 	}
				//limpar waiting room
			}

			// Começa a receber mensagem do cliente
			for {
				length, err := con.Read(data)
				if err != nil {
					fmt.Printf("O cliente %s saiu.\n", name)
					con.Close()
					disconnect(con, name)
					return
				}
				res = string(data[:length])
				fmt.Println(res)
				handleCommand(res)
			}
		}(conn)
	}
}

func handleCommand(clientMessage string) {
	command := strings.Split(clientMessage, ":")
	if len(command) > 1 {
		command[len(command)-1] = strings.Replace(command[len(command)-1], "\n", "", -1)
	}
	//Sempre colocar \n
	//Não usar : quando não for outro comando
	fmt.Println("Client Comando : " + command[0])
	switch command[0] {
	case "masterAnswer":
		answer := "serverMessage:O mestre respondeu ' " + command[1] + " '\n"
		fmt.Println("O mestre respondeu : " + command[1])
		notifyAll(answer, waitingRoom)
		notifyPlayer("yourTurnTry\n", waitingRoom[0])
		//setar o próximo a perguntar

	case "setGuess":
		guess = strings.ToUpper(command[1]) //Disable case sensitive answers
		fmt.Println("O mestre definiu o seguinte personagem : " + guess)
	case "sendFirstTip":
		tip = "serverMessage:A dica concedida pelo mestre é " + command[1] + "\n"
		fmt.Println(tip)
		notifyAllButYou(waitingRoom[1], tip, waitingRoom)
		notifyPlayer("yourTurnAsk\n", waitingRoom[0]) //Primeiro a perguntar
	case "sendQuestion":
		question := "serverMessage:O jogador [" + command[1] + "] perguntou ao mestre ' " + command[2] + " '\n"
		fmt.Println(question)
		notifyAll(question, waitingRoom)
		notifyPlayer("ansAsMaster:"+command[1]+"\n", waitingRoom[1]) //Pede resposta ao mestre
	case "tryGuess":
		attempt := strings.ToUpper(command[1])
		if guess == attempt {
			fmt.Println("O JOGADOR GANHOU!")
		}
	}
}

// Notifica um player especifico
func notifyPlayer(msg string, receiver net.Conn) {
	receiver.Write([]byte(msg))
}

// Notifica todos os players da lista
func notifyAll(msg string, receivers []net.Conn) {
	for _, con := range receivers {
		con.Write([]byte(msg))
	}
}

// Notifica todos os players da lista exceto um
func notifyAllButYou(conn net.Conn, msg string, receivers []net.Conn) {
	for _, con := range receivers {
		if con.RemoteAddr() != conn.RemoteAddr() {
			con.Write([]byte(msg))
		}
	}
}

// Deleta os players desconectados e notifica os outros
func disconnect(conn net.Conn, name string) {
	for index, con := range clients {
		if con.RemoteAddr() == conn.RemoteAddr() {
			disMsg := name + " deixou a partida."
			fmt.Println(disMsg)
			clients = append(clients[:index], clients[index+1:]...)
			notifyAll(disMsg, waitingRoom)
		}
	}
}
