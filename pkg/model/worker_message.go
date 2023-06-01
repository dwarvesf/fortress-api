package model

type Action uint8

type WorkerMessage struct {
	Type    string
	Payload interface{}
}
