package device

type SerialPort struct {
 Address string `json:"address"`
 Label string `json:"label,omitempty"`
 Protocol string `json:"protocol,omitempty"`
 ProtocolLabel string `json:"protocol_label,omitempty"`
 MatchingBoards []BoardMatch `json:"matching_boards,omitempty"`
}
type BoardMatch struct {
 FQBN string `json:"fqbn,omitempty"`
 Name string `json:"name,omitempty"`
}
