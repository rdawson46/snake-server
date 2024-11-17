package packet

import (
    "encoding/json"
)

type Packet struct {
    Version string `json:"version"`
    Length  int    `json:"length"`
    Width   int    `json:"width"`
    Page    string `json:"page"`
}

func Encode(p Packet) ([]byte, error) {
    return json.Marshal(p)
}

func Decode(b []byte) (*Packet, error) {
    p := &Packet{}
    err := json.Unmarshal(b, p)
    return p, err
}

