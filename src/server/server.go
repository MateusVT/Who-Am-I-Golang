package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var clients []net.Conn
var matchs []Match
var playerNames = ""
var waitingRoom []net.Conn
var maxPlayersPerMatch = 3
var match Match
var guess string
var tip string
var idTurnPlayer int

type Match struct {
	matchID        int
	playersCon     []net.Conn
	players        []Player
	idMasterPlayer int
	guess          string //Defined by the Master
	tip            string //Defined by the Master
	// idTurnPlayer   int
	winner *Player
	ended  bool

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
		match.playersCon = append(match.playersCon, conn)

		go func(con net.Conn) {

			// con.Write([]byte(msg))

			length, err := con.Read(data)
			name := string(data[:length])
			fmt.Println("Nova conexão : ", con.RemoteAddr())
			fmt.Println("O Jogador : ", name, "entrou no servidor!")
			if err != nil {
				fmt.Printf("O cliente %v saiu.\n", con.RemoteAddr())
				// for _, conMatch := range match.playersCon {
				// 	if conMatch.RemoteAddr() == con.RemoteAddr() {

				// 	}
				// }
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

			// min := 0
			// max := 2

			player := Player{}
			player.con = con
			player.name = name
			player.score = 0
			match.players = append(match.players, player)

			// match.idMasterPlayer = rand.Intn(max-min) + min
			match.idMasterPlayer = 0
			match.ended = false

			notifyAllButYou(con, comeStr, waitingRoom)

			if len(waitingRoom) < maxPlayersPerMatch {
				// notifyPlayer("serverMessage:"+"Atualmente há: "+strconv.Itoa(len(clients))+" jogadores conectados no servidor.\n", con)
				// <-timer.C
				notifyPlayer("serverMessage: "+strconv.Itoa(len(waitingRoom))+" jogadores na sala de espera.\n", con)

			}

			if len(waitingRoom) == maxPlayersPerMatch {

				notifyAll("initializeMatch:1:"+strconv.Itoa(len(waitingRoom))+":"+playerNames+":"+strings.Split(playerNames, ",")[0]+"\n", match.playersCon)
				// <-timer.C
				// notifyAll("initializeMatch:1", waitingRoom)
				notifyPlayer("setMaster\n", match.playersCon[match.idMasterPlayer])
				notifyAllButYou(match.playersCon[match.idMasterPlayer], "serverMessage:O mestre está definindo o personagem a ser advinhado, aguarde sua vez.\n", match.playersCon)
				// match.idTurnPlayer = 1
				idTurnPlayer = 1

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

				// match.playersCon
				length, err := con.Read(data)
				if err != nil {
					fmt.Printf("O cliente %s saiu.\n", name)
					con.Close()
					disconnect(con, name)
					return
				}
				res = string(data[:length])
				// fmt.Println("msg res:", res)

				// fmt.Println("Turno do cara: ", match.idTurnPlayer)

				if res == "showRanking" {
					handleRequestCommand(res, match, con)
				} else {
					handleCommand(res, match)

				}
			}
		}(conn)
	}
}

func handleCommand(clientMessage string, match Match) {

	command := strings.Split(clientMessage, ":")
	if len(command) > 1 {
		command[len(command)-1] = strings.Replace(command[len(command)-1], "\n", "", -1)
	}
	// match.idMasterPlayer
	//Sempre colocar \n
	//Não usar : quando não for outro comando
	// fmt.Println("Client Comando : " + command[0])

	switch command[0] {
	case "masterAnswer":
		answer := "serverMessage:O mestre respondeu ' " + command[1] + " '\n"
		fmt.Println("O mestre respondeu : " + command[1])
		// notifyAll(answer, match.playersCon)
		notifyAllButYou(match.playersCon[match.idMasterPlayer], answer, match.playersCon)
		fmt.Printf("match.idTurnPlayer after master answaer %v", idTurnPlayer)
		notifyPlayer("yourTurnTry\n", match.playersCon[idTurnPlayer]) //Aqui ta sempre 1, não ta subindo.
		//setar o próximo a perguntar

	case "setGuess":
		guess = strings.ToUpper(command[1]) //Disable case sensitive answers
		fmt.Println("O mestre definiu o seguinte personagem : " + guess)
	case "sendFirstTip":
		tip = "serverMessage:A dica concedida pelo mestre é '" + command[1] + "'\n"
		fmt.Println(tip)
		println(match.idMasterPlayer)
		println(match.players[match.idMasterPlayer].name)

		notifyAllButYou(match.playersCon[match.idMasterPlayer], tip, match.playersCon)
		notifyPlayer("yourTurnAsk\n", match.playersCon[idTurnPlayer]) //Primeiro a perguntar
	case "sendQuestion":
		question := "serverMessage:O jogador [" + command[1] + "] perguntou ao mestre ' " + command[2] + " '\n"
		fmt.Println(question)
		notifyAllButYou(match.playersCon[idTurnPlayer], question, match.playersCon)
		notifyPlayer("ansAsMaster:"+command[1]+"\n", match.playersCon[match.idMasterPlayer]) //Pede resposta ao mestre
	case "tryGuess":
		attempt := strings.ToUpper(command[1])
		if guess == attempt {
			match.ended = true
			fmt.Println("O JOGADOR GANHOU!")
			match.players[idTurnPlayer].score = 10000
			err := escreverTexto(match)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				fmt.Println("Arquivo salvo com sucesso.")
			}
			notifyAll("O player "+match.players[idTurnPlayer].name+" acertou!!!\n", match.playersCon)

		} else {
			notifyPlayer("serverMessage:Resposta errada! Aguarde sua próxima tentativa.\n", match.playersCon[idTurnPlayer])
			notifyAllButYou(match.playersCon[idTurnPlayer], "serverMessage:O jogador["+match.players[idTurnPlayer].name+"] chutou '"+attempt+"' e errou.\n", match.playersCon)
			idTurnPlayer = idTurnPlayer + 1 //Aqui ta somando mas não ta batendo lá
			if idTurnPlayer == maxPlayersPerMatch {
				idTurnPlayer = 1
				fmt.Printf("Entrei no if? %v", idTurnPlayer)
			}
			notifyPlayer("yourTurnAsk\n", match.playersCon[idTurnPlayer]) //Primeiro a perguntar
		}

	}

}

func handleRequestCommand(clientMessage string, match Match, sender net.Conn) {

	command := strings.Split(clientMessage, ":")
	if len(command) > 1 {
		command[len(command)-1] = strings.Replace(command[len(command)-1], "\n", "", -1)
	}

	switch command[0] {

	case "showRanking":
		notifyPlayer("serverMessage:Fulano - 2000 pontos\n", sender)

	}

}

func escreverTexto(match Match) error {
	// Cria o arquivo de texto
	arquivo, err := os.OpenFile("/parser.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// Caso tenha encontrado algum erro retornar ele
	if err != nil {
		return err
	}
	// Garante que o arquivo sera fechado apos o uso
	defer arquivo.Close()

	var conteudo []string
	conteudo = append(conteudo, "-\n")
	conteudo = append(conteudo, "Partida encerrada, matchId: "+strconv.Itoa(match.matchID)+"\n")
	conteudo = append(conteudo, "- \n")
	conteudo = append(conteudo, "Vencedor: "+match.players[idTurnPlayer].name+"\n"+" Score: "+strconv.Itoa(match.players[idTurnPlayer].score))
	// Cria um escritor responsavel por escrever cada linha do slice no arquivo de texto
	escritor := bufio.NewWriter(arquivo)
	for _, linha := range conteudo {
		fmt.Fprintln(escritor, linha)
	}

	// Garante que o arquivo sera fechado apos o uso
	defer arquivo.Close()
	// Caso a funcao flush retorne um erro ele sera retornado aqui tambem
	return escritor.Flush()
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
			// fmt.Println(disMsg)
			clients = append(clients[:index], clients[index+1:]...)
			notifyAll(disMsg, waitingRoom)
		}
	}
}
