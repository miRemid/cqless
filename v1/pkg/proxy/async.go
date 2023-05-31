package proxy

import (
	"bytes"
	"encoding/gob"
	"net/http"

	"github.com/nats-io/nats.go"
)

type AsyncRequest struct {
	RawRequest *http.Request
	FuncName   string
	Callback   string
}

type AsyncCallbackRequest struct {
	Status int
	Body   []byte
}

func (p *Proxy) asyncSubChannel() {
	for {
		select {
		case async := <-p.asyncQueue:
			p.asyncSend(async)
		case <-p.asyncClose:
			return
		}
	}
}

func (p *Proxy) asyncSubNats(msg *nats.Msg) {
	evt := p.log.With().Str("action", "async-request").Logger()
	buffer := bytes.NewBuffer(msg.Data)
	var async = &AsyncRequest{}
	if err := gob.NewDecoder(buffer).Decode(&async); err != nil {
		evt.Err(err).Msg("decode msg error!")
		return
	}
	p.asyncSend(async)
}

func (p *Proxy) asyncSend(async *AsyncRequest) {
	evt := p.log.With().Str("action", "async-request").Logger()
	resp, err := p.send(async.RawRequest)
	if err != nil {
		evt.Err(err).Msg("invoke function failed")
		return
	}
	defer resp.Body.Close()
	callbackReq, err := http.NewRequest(http.MethodPost, async.Callback, resp.Body)
	if err != nil {
		evt.Err(err).Str("callback", async.Callback).Msg("invalid callback")
		return
	}
	callbackResp, err := p.send(callbackReq)
	if err != nil {
		evt.Err(err).Str("callback", async.Callback).Msg("invoke callback error")
		return
	}
	if callbackResp.StatusCode < 200 || callbackResp.StatusCode >= 300 {
		evt.Error().Str("callback", async.Callback).Msg("invoke callback error")
		return
	}
}
