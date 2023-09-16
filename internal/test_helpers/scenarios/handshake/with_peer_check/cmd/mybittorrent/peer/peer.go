package peer

import "fmt"

const IpLength = 4
const PortLength = 2
const BytesLength = IpLength + PortLength

type Peer struct {
	Ip   string
	Port int
}

func NewPeer(bytes []byte) (*Peer, error) {
	if len(bytes) != BytesLength {
		return nil, fmt.Errorf("Peers are expected to have %d bytes", BytesLength)
	}
	ipBytes := bytes[:IpLength]
	portBytes := bytes[IpLength : IpLength+PortLength]

	temp := int(portBytes[0])
	port := int((temp << 8) | int(portBytes[1]))
	return &Peer{
		Ip:   fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]),
		Port: port,
	}, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("%s:%d", p.Ip, p.Port)
}
