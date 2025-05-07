package superobject

import "fmt"

func HandleErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
}