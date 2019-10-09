package hashring

import (
	"fmt"
)

func ExampleHashRing() {
	hr := new(HashRing).Init()

	hr.AddNode("192.168.1.1", 100)
	hr.AddNode("192.168.1.2", 200)
	hr.AddNode("192.168.1.3", 300)

	s1, _ := hr.FindNode("user_a")
	s2, _ := hr.FindNode("user_b")
	s3, _ := hr.FindNode("user_c")
	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(s3)
	// Output:
	// 192.168.1.3
	// 192.168.1.3
	// 192.168.1.2
}
