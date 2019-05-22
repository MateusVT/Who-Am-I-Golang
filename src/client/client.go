package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	// "github.com/go-delve/delve/pkg/dwarf/reader"
)

var writeStr, readStr = make([]byte, 1024), make([]byte, 1024)

var serverCon net.Conn
var menu = 0
var enterTheGame = false
var (
	reader  = bufio.NewReader(os.Stdin)
	scanner = bufio.NewScanner(os.Stdin)
	laddr   string
	// PlayerScore  int
	name string

	isMaster     bool
	waitingInput = true
	inputType    string
	runningMatch bool
	playerID     int
	matchID      int
	score        = 1000
)

func main() {
	var (
		host   = "127.0.0.1"
		port   = "8000"
		remote = host + ":" + port
	)
	timer := time.NewTimer(1 * time.Second)
	serverCon, err := net.Dial("tcp", remote)
	defer serverCon.Close()

	if err != nil {
		fmt.Println("Servidor não encontrado!.")
		os.Exit(-1)
	}

	fmt.Printf("----------------------------------------------------------------\n")
	fmt.Printf("BEM VINDO AO 'QUEM SOU EU?'\n")
	fmt.Printf("----------------------------------------------------------------\n")
	fmt.Println()
	fmt.Printf("Digite seu nome para entrar no jogo : ")
	fmt.Scanf("%s", &name)

	sendMessageToServer(name, serverCon)

	fmt.Printf(" _____________________________\n")
	fmt.Printf("|           MENU   		|\n")
	fmt.Printf("|     1 - Exibir Ranking      |\n")
	fmt.Printf("|     2 - Buscar Partida      |\n")
	fmt.Printf("|_____________________________|\n")
	fmt.Println()
	go read(serverCon)
	for {
		if waitingInput {

			if menu == 0 && !runningMatch {
				<-timer.C
				fmt.Printf("Selecione uma opção : \n")
				reader.ReadLine()
				scanner.Scan()
				opt := scanner.Text()
				optConv, err := strconv.Atoi(opt)
				if err == nil {
				}
				menu = optConv
			}

			if menu == 1 {

				sendMessageToServer("showRanking", serverCon)
				// waitingInput = true
				menu = 0
			} else {
				fmt.Printf("Você entrou na sala de espera, aguarde outros jogadores.\n")
				waitingInput = false
			}

			// if waitingInput {

			switch inputType {
			case "masterInit":
				reader.ReadLine()
				scanner.Scan()
				guess := scanner.Text()
				sendMessageToServer("setGuess:"+guess+"\n", serverCon)
				fmt.Println("Informe a primeira dica : ")
				scanner.Scan() // use `for scanner.Scan()` to keep reading
				tip := scanner.Text()
				sendMessageToServer("sendFirstTip:"+tip+"\n", serverCon)
				fmt.Println("O jogador da vez está elaborando uma pergunta, aguarde.")

				waitingInput = false
			case "askQuestion":
				reader.ReadLine()
				scanner.Scan()
				question := scanner.Text()
				fmt.Println("Aguarde a resposta do Mestre : ")
				sendMessageToServer("sendQuestion:"+name+":"+question+"\n", serverCon)
				waitingInput = false
			case "ansQuestion":
				// reader.ReadLine()
				scanner.Scan()
				answer := scanner.Text()
				fmt.Println("Sua resposta foi enviada para os outros jogadores!")
				fmt.Println("Aguarde a próxima pergunta.")
				sendMessageToServer("masterAnswer:"+answer+"\n", serverCon)
				waitingInput = false

			case "tryGuess":
				scanner.Scan()
				trial := scanner.Text()
				sendMessageToServer("tryGuess:"+trial+"\n", serverCon)
				waitingInput = false
			default:
			}
		}
	}
}

func handleCommand(serverMessage string) {
	command := strings.Split(serverMessage, ":")
	// fmt.Println("Server Original Message : " + serverMessage)
	if len(command) > 1 {
		command[len(command)-1] = strings.Replace(command[len(command)-1], "\n", "", -1)
	}
	// fmt.Println("Server Comando : " + command[0])
	switch command[0] {
	case "initializeMatch":
		runningMatch = true
		id, err := strconv.Atoi(command[1])
		qtyPlayers, err := strconv.Atoi(command[2])
		namePlayers := strings.Split(command[3], ",")
		masterName := command[4]
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		matchID = id
		fmt.Println("----------------------------------------------------------------")
		fmt.Println(" PARTIDA INICIADA")
		fmt.Println("----------------------------------------------------------------")
		fmt.Printf("Jogadores conectados: %v \n", qtyPlayers)
		for i := 0; i < len(namePlayers); i++ {
			fmt.Println("	[" + namePlayers[i] + "]")
		}
		fmt.Println("MESTRE da rodada: [" + masterName + "]")

	case "serverMessage":
		fmt.Println(command[1])

	case "setMaster":
		isMaster = true
		// fmt.Println("Você foi selecionado para ser o Mestre dessa partida!")
		fmt.Println("----------------------------------------------------------------")
		fmt.Println(" VOCÊ FOI SELECIONADO COMO MESTRE DESTA PARTIDA")
		fmt.Println("----------------------------------------------------------------")
		fmt.Println("Informe o personagem a ser adivinhado pelos outros jogadores : ")
		inputType = "masterInit"
		waitingInput = true

	case "ansAsMaster":
		fmt.Println("Responda a pergunta do jogador [" + command[1] + "] - (SIM ou NÃO)")
		inputType = "ansQuestion"
		waitingInput = true

	case "setPlayerID":
		id, err := strconv.Atoi(command[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		playerID = id

	case "yourTurnAsk":
		fmt.Println("Seu turno!")
		fmt.Println("Pergunte algo ao Mestre : ")
		inputType = "askQuestion"

		waitingInput = true

	case "yourTurnTry":
		fmt.Println("Sua tentativa :")
		inputType = "tryGuess"
		waitingInput = true

	case "decreaseScore":
		subScore, err := strconv.Atoi(command[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		score = score - subScore
		fmt.Println("decreseScore : " + strconv.Itoa((subScore)))
		fmt.Println("Score : " + strconv.Itoa((score)))
	default:
		fmt.Println(command[0])
	}
}

// Notify all other clients
func sendMessageToServer(msg string, receiver net.Conn) {
	in, err := receiver.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Erro ao enviar mensagem para o servidor: %d\n", in)
		os.Exit(0)
	}

}

func read(conn net.Conn) {
	for {
		length, err := conn.Read(readStr)
		if err != nil {
			fmt.Printf("Error when read from server. Error:%s\n", err)
			os.Exit(0)
		}

		var serverMessage string
		serverMessage = string(readStr[:length])
		message := strings.Split(serverMessage, "\n")
		for i := 0; i < len(message)-1; i++ {
			handleCommand(message[i])
		}

	}
}
