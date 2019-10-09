package hashring

import (
	"fmt"
	"math"
	"testing"
)

func TestHashRing(t *testing.T) {
	hr := new(HashRing).Init()

	_, ok := hr.FindNode("foo")
	if ok {
		t.Fatal()
	}

	// ----------

	ok = hr.AddNode("bar", 1000)
	if !ok {
		t.Fatal()
	}
	v, ok := hr.FindNode("foo")
	if !ok {
		t.Fatal()
	}
	if v != "bar" {
		t.Fatal(v)
	}

	// ----------

	ok = hr.AddNode("bar", 1000)
	if ok {
		t.Fatal()
	}

	ok = hr.AddNode("bar2", 2000)
	if !ok {
		t.Fatal()
	}
	ok = hr.AddNode("bar3", 3000)
	if !ok {
		t.Fatal()
	}

	// ----------

	m := map[string]int{
		"bar":  0,
		"bar2": 0,
		"bar3": 0,
	}
	for i := 0; i < 600000; i++ {
		v, ok = hr.FindNode(fmt.Sprintf("foo%d", i))
		if !ok {
			t.Fatal()
		}
		m[v]++
	}
	if len(m) != 3 {
		t.Fatal(m)
	}
	if math.Abs(float64(m["bar"]-100000)) > 10000 {
		t.Fatal(m)
	}
	if math.Abs(float64(m["bar2"]-200000)) > 20000 {
		t.Fatal(m)
	}
	if math.Abs(float64(m["bar3"]-300000)) > 30000 {
		t.Fatal(m)
	}

	// ----------

	hr.RemoveNode("bar2")
	m = map[string]int{
		"bar":  0,
		"bar3": 0,
	}
	for i := 0; i < 400000; i++ {
		v, ok = hr.FindNode(fmt.Sprintf("foo%d", i))
		if !ok {
			t.Fatal()
		}
		m[v]++
	}
	if len(m) != 2 {
		t.Fatal(m)
	}
	if math.Abs(float64(m["bar"]-100000)) > 10000 {
		t.Fatal(m)
	}
	if math.Abs(float64(m["bar3"]-300000)) > 30000 {
		t.Fatal(m)
	}

	// ----------

	hr.RemoveNode("bar3")
	for i := 0; i < 100000; i++ {
		v, ok = hr.FindNode(fmt.Sprintf("foo%d", i))
		if !ok {
			t.Fatal()
		}
		if v != "bar" {
			t.Fatal(m, v)
		}
	}

	// ----------

	hr.RemoveNode("bar")
	for i := 0; i < 100000; i++ {
		v, ok = hr.FindNode(fmt.Sprintf("foo%d", i))
		if ok {
			t.Fatal()
		}
	}
}
