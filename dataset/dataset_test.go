package dataset

import (
	"log"
	"os"
	"testing"
)

func TestGenTestData(t *testing.T) {
	os.Setenv("PROJECT_ROOT", "/home/ubuntu/contractdb")
	//GenTpchWorkloadQ6(100)
	//for _, ni := range []int{14, 15, 16, 17, 18, 19, 20} {
	for _, ni := range []int{10} {
		GenTestData(ni)
		log.Printf("generate workload for N=2^%v ok\n", ni)
	}
}
