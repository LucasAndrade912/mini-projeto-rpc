package main

import (
	"bufio"
	"fmt"
	remoteList "mini-projeto-rpc/remotelist/pkg"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

func main() {
	client, err := rpc.Dial("tcp", "rpc-server:5000")

	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer client.Close()

	fmt.Println("=== Cliente RPC Interativo ===")
	fmt.Println("Conectado ao servidor RPC em :5000")
	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  append <list_id> <value>  - Adiciona um valor à lista")
	fmt.Println("  get <list_id> <index>     - Obtém valor em uma posição")
	fmt.Println("  remove <list_id>          - Remove último elemento da lista")
	fmt.Println("  size <list_id>            - Obtém tamanho da lista")
	fmt.Println("  exit                      - Sair do programa")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("rpc> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		switch command {
		case "exit", "quit":
			fmt.Println("Encerrando cliente...")
			return

		case "append":
			handleAppend(client, parts)

		case "get":
			handleGet(client, parts)

		case "remove":
			handleRemove(client, parts)

		case "size":
			handleSize(client, parts)

		default:
			fmt.Printf("Comando desconhecido: %s\n", command)
			fmt.Println("Digite 'help' para ver os comandos disponíveis.")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Erro ao ler entrada: %v\n", err)
	}
}

func handleAppend(client *rpc.Client, parts []string) {
	if len(parts) != 3 {
		fmt.Println("Uso: append <list_id> <value>")
		fmt.Println("Exemplo: append 1 42")
		return
	}

	listID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("ID da lista inválido: %s\n", parts[1])
		return
	}

	value, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("Valor inválido: %s\n", parts[2])
		return
	}

	args := remoteList.AppendArgs{
		List_ID: listID,
		Value:   value,
	}

	var reply bool
	err = client.Call("RemoteList.Append", args, &reply)
	if err != nil {
		fmt.Printf("Erro ao chamar Append: %v\n", err)
		return
	}

	if reply {
		fmt.Printf("✓ Valor %d adicionado à lista %d\n", value, listID)
	} else {
		fmt.Printf("✗ Falha ao adicionar valor %d à lista %d\n", value, listID)
	}
}

func handleGet(client *rpc.Client, parts []string) {
	if len(parts) != 3 {
		fmt.Println("Uso: get <list_id> <index>")
		fmt.Println("Exemplo: get 1 0")
		return
	}

	listID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("ID da lista inválido: %s\n", parts[1])
		return
	}

	index, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("Índice inválido: %s\n", parts[2])
		return
	}

	args := remoteList.GetArgs{
		List_ID: listID,
		Index:   index,
	}

	var reply int
	err = client.Call("RemoteList.Get", args, &reply)
	if err != nil {
		fmt.Printf("Erro ao chamar Get: %v\n", err)
		return
	}

	fmt.Printf("Valor na posição %d da lista %d: %d\n", index, listID, reply)
}

func handleRemove(client *rpc.Client, parts []string) {
	if len(parts) != 2 {
		fmt.Println("Uso: remove <list_id>")
		fmt.Println("Exemplo: remove 1")
		return
	}

	listID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("ID da lista inválido: %s\n", parts[1])
		return
	}

	var reply int
	err = client.Call("RemoteList.Remove", listID, &reply)
	if err != nil {
		fmt.Printf("Erro ao chamar Remove: %v\n", err)
		return
	}

	fmt.Printf("✓ Elemento removido da lista %d: %d\n", listID, reply)
}

func handleSize(client *rpc.Client, parts []string) {
	if len(parts) != 2 {
		fmt.Println("Uso: size <list_id>")
		fmt.Println("Exemplo: size 1")
		return
	}

	listID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("ID da lista inválido: %s\n", parts[1])
		return
	}

	var reply int
	err = client.Call("RemoteList.Size", listID, &reply)
	if err != nil {
		fmt.Printf("Erro ao chamar Size: %v\n", err)
		return
	}

	fmt.Printf("Tamanho da lista %d: %d elementos\n", listID, reply)
}
