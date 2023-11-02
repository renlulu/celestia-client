package main

import (
	"fmt"

	blobtypes "github.com/celestiaorg/celestia-app/x/blob/types"
)

func main() {
	gasLimit := blobtypes.DefaultEstimateGas([]uint32{1})
	fmt.Println(gasLimit)
}
