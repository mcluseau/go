package streamsse

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gomodules.xyz/jsonpatch/v2"

	"m.cluseau.fr/go/watchable"
)

func Stream[T any](w http.ResponseWriter, req *http.Request, wable *watchable.Watchable[T]) {
	type watchUpdate struct {
		Set   json.RawMessage       `json:"set,omitempty"`
		Patch []jsonpatch.Operation `json:"p,omitempty"`
		Err   string                `json:"err,omitempty"`
	}

	tickInterval := time.Second / 20 // 20 FPS by default
	if reqInterval := req.FormValue("tick"); reqInterval != "" {
		tickInterval, err := time.ParseDuration(reqInterval)
		if err != nil {
			http.Error(w, "invalid tick: "+err.Error(), http.StatusBadRequest)
			return
		}

		const minInterval = 10 * time.Millisecond
		if tickInterval < minInterval {
			http.Error(w, "tick below min ("+minInterval.String()+"): "+reqInterval, http.StatusBadRequest)
			return
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	send := func(update watchUpdate) (ok bool) {
		ba, err := json.Marshal(update)
		if err != nil {
			log.Print("WARNING: failed to marshal update: ", err)
			return
		}

		_, err = w.Write([]byte("data: " + string(ba) + "\n\n"))
		if err != nil {
			log.Print("update send error: ", err)
			return
		}

		flusher.Flush()

		ok = true
		return
	}

	watch, watchStop := wable.NewWatchCh()
	defer watchStop()

	currentValue, ok := <-watch
	if !ok {
		send(watchUpdate{Err: "watch closed before any value was set"})
		return
	}

	var prevBytes []byte
	{
		ba, err := json.Marshal(currentValue)
		if err != nil {
			log.Print("WARNING: failed to marshal value, failing: ", err)
			send(watchUpdate{Err: "marshal error: " + err.Error()})
			return
		}

		if !send(watchUpdate{Set: ba}) {
			return
		}

		prevBytes = ba
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	gotUpdate := false

	for {
		select {
		case <-req.Context().Done():
			return

		case currentValue = <-watch:
			gotUpdate = true

		case <-ticker.C:
			if !gotUpdate {
				break
			}
			gotUpdate = false

			ba, err := json.Marshal(currentValue)
			if err != nil {
				log.Print("WARNING: failed to marshal value, failing: ", err)
				send(watchUpdate{Err: "marshal error: " + err.Error()})
				return
			}

			if bytes.Equal(prevBytes, ba) {
				break
			}

			patch, err := jsonpatch.CreatePatch(prevBytes, ba)
			if err != nil {
				log.Print("WARNING: failed to compute patch, failing: ", err)
				send(watchUpdate{Err: "compute patch error: " + err.Error()})
				return
			}

			if !send(watchUpdate{Patch: patch}) {
				return
			}

			prevBytes = ba
		}
	}
}
