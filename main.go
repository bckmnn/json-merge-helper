package main

import (
	"fmt"
	"os"

	"github.com/bckmnn/json-merge-helper/sgjsonformat"
)

func main() {

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) != 3 {
		fmt.Println("merge-driver needs 3 arguments")
	} else {
		ancestor := argsWithoutProg[0]
		current := argsWithoutProg[1]
		other := argsWithoutProg[2]

		fmt.Println(ancestor)
		fmt.Println(current)
		fmt.Println(other)

		fmt.Printf("ancestor: %s\n", ancestor)
		ancestorJson := sgjsonformat.NewSgJsonFile(ancestor)
		err := ancestorJson.Read()
		if err != nil {
			fmt.Printf("[Error] %v\n", err)
		}

		fmt.Printf("current: %s\n", current)
		currentJson := sgjsonformat.NewSgJsonFile(current)
		err = currentJson.Read()
		if err != nil {
			fmt.Printf("[Error] %v\n", err)
		}

		fmt.Printf("other: %s\n", other)
		otherJson := sgjsonformat.NewSgJsonFile(other)
		err = otherJson.Read()
		if err != nil {
			fmt.Printf("[Error] %v\n", err)
		}

		combiendIds := ancestorJson.Ids
		combiendIds = append(combiendIds, currentJson.Ids...)
		combiendIds = append(combiendIds, otherJson.Ids...)
		allIds := sgjsonformat.RemoveDuplicates(combiendIds)

		for _, id := range allIds {
			currentE := currentJson.ById[id]
			otherE := otherJson.ById[id]
			currentE.Compare(&otherE)
		}
	}

}
