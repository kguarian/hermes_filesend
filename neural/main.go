package main

import "fmt"

func main() {
	var errCode int
	var nn *NeuralNet

	nn, errCode = buildPerceptron([]int{9, 10, 9, 10})
	if errCode != 0 {
		println("perceptron build error")
		return
	}
	fmt.Printf("%p\n", nn)
}
