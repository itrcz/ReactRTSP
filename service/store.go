package main

import (
	"sync"

	"github.com/deepch/vdk/av"
)

var Store = storeInit()

type StoreST struct {
	mutex   sync.RWMutex
	Streams map[string]StreamST
}

type StreamST struct {
	Active  bool
	Codecs  []av.CodecData
	Clients map[string]ClientST
}

type ClientST struct {
	c chan av.Packet
}

func storeInit() *StoreST {
	var cfg StoreST
	cfg.Streams = make(map[string]StreamST)
	return &cfg
}
