import React, { useEffect, useRef, useState } from 'react'
import { Button, Flexbox, Grid, TextField, Viewport } from '@stage-ui/core'
import { render } from 'react-dom'

let ws: WebSocket

function App() {
  const video = useRef<HTMLVideoElement>(null)
  const [streaming, setStreaming] = useState(false)

  const [state, setState] = useState<Record<string, string>>({
    Host: 'wowzaec2demo.streamlock.net',
    Port: '80',
    Uri: '/vod/mp4:BigBuckBunny_115k.mov',
    Login: '',
    Password: '',
  })

  const stop = () => {
    if (video.current) {
      video.current.src = ''
    }
    ws?.close()
  }
  const start = () => {
    if (video.current) {
      let sourceBuffer: SourceBuffer | null
      let streamingStarted = false
      let mediaSource = new MediaSource()
      let queue: BufferSource[] = []

      const sourceOpen = () => {
        ws = new WebSocket(`ws://${window.location.hostname}:3333/stream?protocol=rtsp&login=${state.Login}&password=${state.Password}&host=${state.Host}&port=${state.Port}&uri=${state.Uri}`)
        ws.binaryType = "arraybuffer"
        ws.onopen = () => {
          setStreaming(true)
          console.log(`Stream on!`)
        }
        ws.onclose = () => {
          setStreaming(false)
          console.log(`Stream off!`)
        }
        ws.onerror = (error) => console.log(error)
        ws.onmessage = (event) => {
          const data = new Uint8Array(event.data)
          if (data[0] != 9) {
            return pckPush(event.data)
          }
          let codec
          const decoded = data.slice(1)
          codec = new TextDecoder("utf-8").decode(decoded)
          sourceBuffer = mediaSource.addSourceBuffer('video/mp4; codecs="' + codec + '"')
          sourceBuffer.mode = "segments"
          sourceBuffer.addEventListener("updateend", pckLoad)
        }
      }

      const pckLoad = () => {
        if (sourceBuffer) {
          if (!sourceBuffer.updating && queue.length > 0) {
            return sourceBuffer.appendBuffer(queue.shift()!)
          }
          streamingStarted = false
        }
      }

      const pckPush = (bufferSource: BufferSource) => {
        if (!sourceBuffer) {
          return
        }
        if (!streamingStarted) {
          sourceBuffer.appendBuffer(bufferSource)
          streamingStarted = true
          return
        }
        queue.push(bufferSource)
        pckLoad()
      }

      mediaSource.addEventListener('sourceopen', sourceOpen, false)
      video.current.src = window.URL.createObjectURL(mediaSource)
    }
  }

  useEffect(() => {
    start()
    return () => stop()
  }, [])

  return (
    <Viewport>
      <Flexbox column centered m="m">
        <video controls muted ref={video} style={{ width: '40rem' }} />
        <Grid mt="m" gap="1rem" templateColumns="1fr 1fr 1fr">
          {Object.keys(state).map((key) => (
            <TextField
              key={key}
              placeholder={key}
              value={state[key]}
              onChange={({ target }) => setState({ ...state, [key]: target.value })}
            />
          ))}
          <Button
            label={streaming ? 'Stop video stream' : 'Start video stream'}
            color={streaming ? 'error' : 'primary'}
            onClick={() => {
              streaming ? stop() : start()
            }}
          />
        </Grid>
      </Flexbox>
    </Viewport>
  )
}

render(<App />, document.getElementById('app'))
