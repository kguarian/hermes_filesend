package main

var ActivationFunction func(tgt Neuron) (ret bool, lastClock uint64) = __activate

type Neuron struct {
	threshold          float64
	feeds              []*Neuron
	weights            []float64
	localControl       *Neuron
	execControl        *Neuron
	clock              *clock
	activationFunction *func(tgt Neuron) (ret bool, lastClock uint64)
	output             bool
	lastClock          uint64
}

type NeuralNet struct {
	head    *Neuron
	tail    *Neuron
	neurons []*Neuron
}

type clock struct {
	signal uint64
}

func (c clock) Increment() uint64 {
	c.signal = c.signal + 1
	return c.signal
}

func buildPerceptron(layers []int) (retval *NeuralNet, errcode int) {
	var neuronCount, i, j, n int
	var neurons []*Neuron
	var head, tail *Neuron
	var activeClock *clock
	var srcIdLo, srcIdHi, layerId, layerCount, layerDepth, currentNeuronIndex int

	if len(layers) == 0 {
		errcode = 1
		return
	}
	for _, v := range layers {
		if v <= 0 {
			errcode = 1
			return
		}
		//useful in neuron initialization
		neuronCount += v
	}

	layerCount = len(layers)

	activeClock = &clock{signal: 0}
	neurons = make([]*Neuron, neuronCount+2)
	//head,tail=&Neuron{}, &Neuron{}
	neurons[0], neurons[neuronCount-1] = &Neuron{}, &Neuron{}

	//fill in all but first and last neurons
	for i = 0; i < neuronCount-2; i++ {
		neurons[i+1] = &Neuron{
			threshold:          1, //holder values
			localControl:       tail,
			execControl:        head,
			clock:              activeClock,
			activationFunction: &ActivationFunction,
		}
	}

	neurons[0] = head
	neurons[neuronCount-1] = tail
	/**
	Algorithm description:
		Motivation: we store the neurons in a 1d array, but the neurons should be properly interconnected.
		Goal: to arrange the layers as each others' feeds, where appropriate.
		Specification:
			1) start at the head node
			2) set head as a feed for all the first-layer neurons
	*/
	//n is the lowest index in `neurons` corresponding to a
	n, j = 1, 0
	for i = 0; i < len(layers); i++ {
		n += layers[i]
	}

	return

}

func __activate(tgt Neuron) (retVal bool, lastClock uint64) {
	var i, n int
	var tally float64
	retVal = false
	n = len(tgt.feeds)
	if n == 0 || len(tgt.weights) < n {
		return
	}
	for i = 0; i < n; i++ {
		switch tgt.feeds[i].output {
		case false:
			break
		case true:
			tally += tgt.weights[i]
		}
	}
	tgt.output = (tally >= tgt.threshold)
	lastClock = tgt.clock.signal
	return
}
