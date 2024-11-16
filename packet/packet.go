package packet

import (
    "encoding/json"
    "github.com/rdawson46/snake-server/server"
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

func MakePacket(s *server.Server, b string) Packet {
    return Packet{
        Version: "0.1",
        Length: s.Config.Length,
        Width: s.Config.Width,
        Page: b,
    }
}

