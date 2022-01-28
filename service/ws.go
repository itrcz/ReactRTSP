package main

import (
	"log"
	"time"

	"golang.org/x/net/websocket"

	"github.com/deepch/vdk/format/mp4f"
)

func ws(ws *websocket.Conn) {
	defer ws.Close()

	var start bool
	var timeLine = make(map[int8]time.Duration)

	url := createURL(ws.Request())

	if Store.streamInit(url).Active == false {
		go streamWatch(url)
	}

	uuid, ch := Store.streamClientAdd(url)
	defer Store.streamClientRemove(uuid)
	codecs := Store.streamCodecGet(url)

	if codecs == nil {
		log.Println("No cedecs found")
		return
	}

	ws.SetWriteDeadline(time.Now().Add(5 * time.Second))

	muxer := mp4f.NewMuxer(nil)
	muxer.WriteHeader(codecs)

	meta, init := muxer.GetInit(codecs)

	websocket.Message.Send(ws, append([]byte{9}, meta...))
	websocket.Message.Send(ws, init)

	go func() {
		for {
			var message string
			if websocket.Message.Receive(ws, &message) != nil {
				ws.Close()
				return
			}
		}
	}()

	streamEnd := time.NewTimer(10 * time.Second)

	for {
		select {
		case <-streamEnd.C:
			log.Println("End of stream")
			return
		case pck := <-ch:
			if pck.IsKeyFrame {
				streamEnd.Reset(10 * time.Second)
				start = true
			}
			if !start {
				continue
			}
			timeLine[pck.Idx] += pck.Duration
			pck.Time = timeLine[pck.Idx]
			ready, buf, _ := muxer.WritePacket(pck, false)
			if ready {
				if ws.SetWriteDeadline(time.Now().Add(10*time.Second)) != nil {
					return
				}
				if websocket.Message.Send(ws, buf) != nil {
					return
				}
			}
		}
	}
}
