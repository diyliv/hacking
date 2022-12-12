package main

type Port struct {
	Banner string `json:"banner"`
	Ports  []int  `json:"open_ports"`
}
