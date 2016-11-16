package potential

/*
Receptor receives external output and is connected to a cell. It has a corresponding output cell.
When this receptor fires, it always fires its cell.

When building a network, the developer should inherit Receptor and make `Value` their own type
(instead of leaving it as `interface{}`).
*/
type Receptor struct {
	/*
	   Value should be a unique piece of information which this receptor represents. In the case
	   of a human eye, *cones* are receptors for location and color. So the `Value` might include
	   a struct with x and y location coordinates, the kind of cone (red, green, or blue). The
	   combinations of these values should be unique.
	*/
	Value interface{}
	/*
	   CellToFire is he normal network cell which this Receptor fires when the receptor is
	   activated.
	*/
	CellToFire *Cell
}

/*
Activate should be used to indicate the value this Receptor represents is present/happening,
and the network should respond accordingly. It fires this cell's action potential.
*/
func (receptor *Receptor) Activate() {
	receptor.CellToFire.FireActionPotential()
}
