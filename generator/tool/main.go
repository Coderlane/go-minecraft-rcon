package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Coderlane/go-minecraft-rcon/generator"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Specify a file to parse.")
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out, err := generator.ParseCommandWithAlias(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(out.String())
	}
}
