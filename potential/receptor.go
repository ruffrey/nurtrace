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
	   The normal network cell that this receptor fires.
	*/
	CellFiree *Cell
}
