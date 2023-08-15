package cc

import "encoding/gob"

func init() {
	gob.Register(ZigCC{})
	gob.Register(MapCC{})
	gob.Register(SingleCWrapper{})
	gob.Register(Cilly{})
	gob.Register(Inject{})
	gob.Register(W2c2{})
}
