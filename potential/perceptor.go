package potential

/*
Perceptor corresponds to a receptor, but on the other side of the network.
*/
type Perceptor struct {
	/*
	   Value is the user supplied real world value that this outlet Perceptor represents.
	*/
	Value interface{}
	/*
	   The normal network cell that leads to this Perceptor.
	*/
	InputCell *Cell
}
