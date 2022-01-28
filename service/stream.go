package main

import (
	"errors"
	"log"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtspv2"
)

var (
	ErrStreamNoVideo    = errors.New("Stream close - no video")
	ErrStreamDisconnect = errors.New("Stream close - RTSP disconnected")
	ErrStreamNoClients  = errors.New("Stream close - no clients left")
)

func (cfg *StoreST) streamInit(url string) (s StreamST) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	stream, ok := cfg.Streams[url]
	if !ok {
		newStream := StreamST{
			Active: false,
		}
		newStream.Clients = make(map[string]ClientST)
		cfg.Streams[url] = newStream
		return newStream
	}
	return stream
}

func (cfg *StoreST) streamClientAdd(url string) (string, chan av.Packet) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	uuid := createUUID()
	ch := make(chan av.Packet, 100)
	cfg.Streams[url].Clients[uuid] = ClientST{c: ch}
	return uuid, ch
}

func (cfg *StoreST) streamClientRemove(uuid string) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	for url := range Store.Streams {
		delete(cfg.Streams[url].Clients, uuid)
	}
}

func (cfg *StoreST) streamClientExists(url string) bool {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	if stream, ok := cfg.Streams[url]; ok && len(stream.Clients) > 0 {
		return true
	}
	return false
}

func (cfg *StoreST) streamCast(url string, pck av.Packet) {
	cfg.mutex.Lock()
	defer cfg.mutex.Unlock()
	for _, v := range cfg.Streams[url].Clients {
		if len(v.c) < cap(v.c) {
			v.c <- pck
		}
	}
}

func (store *StoreST) streamCodecAdd(url string, codecs []av.CodecData) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	t := store.Streams[url]
	t.Codecs = codecs
	store.Streams[url] = t
}

func (cfg *StoreST) streamCodecGet(url string) []av.CodecData {
	for i := 0; i < 100; i++ {
		cfg.mutex.RLock()
		tmp, ok := cfg.Streams[url]
		cfg.mutex.RUnlock()
		if !ok {
			return nil
		}
		if tmp.Codecs != nil {
			return tmp.Codecs
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func (store *StoreST) streamActiveSet(url string, state bool) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	t := store.Streams[url]
	t.Active = state
	store.Streams[url] = t
}

func (store *StoreST) RTSP(url string) error {
	videoTest := time.NewTimer(10 * time.Second)
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              url,
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            false,
	})
	if err != nil {
		return err
	}
	defer RTSPClient.Close()
	if RTSPClient.CodecData != nil {
		Store.streamCodecAdd(url, RTSPClient.CodecData)
	}
	for {
		select {
		case <-videoTest.C:
			return ErrStreamNoVideo
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				Store.streamCodecAdd(url, RTSPClient.CodecData)
			case rtspv2.SignalStreamRTPStop:
				return ErrStreamDisconnect
			}
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if !Store.streamClientExists(url) {
				return ErrStreamNoClients
			}
			if packetAV.IsKeyFrame {
				videoTest.Reset(20 * time.Second)
			}
			Store.streamCast(url, *packetAV)
		}
	}
}

func streamWatch(url string) {
	stream := Store.Streams[url]
	if stream.Active == false {
		log.Println("Watch", url)
		Store.streamActiveSet(url, true)
		err := Store.RTSP(url)
		if err != nil {
			Store.streamActiveSet(url, false)
			log.Println(err)
		}
	}
}
