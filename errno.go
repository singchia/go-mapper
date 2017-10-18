package main

const (
	Succeed           int64 = 0
	SomethingFuckedUp int64 = 40010
	ISMEMBEROperError int64 = 40011
	SADDOperError     int64 = 40012
	SREMOperError     int64 = 40013

	AlreayExists        int64 = 40021
	ContradictionExists int64 = 40022
	EntityNotExists     int64 = 40023
	EntitiesLocked      int64 = 40024
	EntitiesNotEnough   int64 = 40025
	EntitiesNotAssigned int64 = 40026

	SETOperError int64 = 40031
	GETOperError int64 = 40032
	DELOperError int64 = 40033
)
