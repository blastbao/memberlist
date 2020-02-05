package memberlist

import (
	"net"
	"time"
)







// Packet is used to provide some metadata about incoming packets from peers
// over a packet connection, as well as the packet payload.
type Packet struct {

	// Buf has the raw contents of the packet.
	Buf []byte


	// From has the address of the peer. This is an actual net.Addr so we
	// can expose some concrete details about incoming packets.
	From net.Addr


	// Timestamp is the time when the packet was received. This should be
	// taken as close as possible to the actual receipt time to help make an
	// accurate RTT measurement during probes.

	Timestamp time.Time
}





// Transport is used to abstract over communicating with other peers.
// Transport 用于封装了与其它节点的通信逻辑。
//
// The packet interface is assumed to be best-effort and the stream interface is assumed to be reliable.
//
//
type Transport interface {



	// FinalAdvertiseAddr is given the user's configured values (which
	// might be empty) and returns the desired IP and port to advertise to
	// the rest of the cluster.
	//
	//
	// 传入 Config 中的 AdvertiseAddr 和 AdvertisePort ，返回的是用于告知集群内其他成员的本节点 ip 和 端口。
	FinalAdvertiseAddr(ip string, port int) (net.IP, int, error)


	// WriteTo is a packet-oriented interface that fires off the given
	// payload to the given address in a connectionless fashion. This should
	// return a time stamp that's as close as possible to when the packet
	// was transmitted to help make accurate RTT measurements during probes.
	//
	// This is similar to net.PacketConn, though we didn't want to expose
	// that full set of required methods to keep assumptions about the
	// underlying plumbing to a minimum. We also treat the address here as a
	// string, similar to Dial, so it's network neutral, so this usually is
	// in the form of "host:port".
	//
	//
	// 向指定 addr 发送一段数据 b ，返回完成通信的当前时刻，该时刻可以用来判断RTT
	WriteTo(b []byte, addr string) (time.Time, error)


	// PacketCh returns a channel that can be read to receive incoming
	// packets from other peers. How this is set up for listening is left as
	// an exercise for the concrete transport implementations.
	//
	// PacketCh 返回一个 chan ，当每从其他成员收到一个 udp 报文时，将报文封装成
	// *Packet 写入该 chan ，上层逻辑通过该 chan 读取 udp 信息。
	PacketCh() <-chan *Packet


	// DialTimeout is used to create a connection that allows us to perform
	// two-way communication with a peer. This is generally more expensive
	// than packet connections so is used for more infrequent operations
	// such as anti-entropy or fallback probes if the packet-oriented probe failed.
	//
	// 创建一个 TCP 连接给上层使用
	DialTimeout(addr string, timeout time.Duration) (net.Conn, error)


	// StreamCh returns a channel that can be read to handle incoming stream
	// connections from other peers. How this is set up for listening is
	// left as an exercise for the concrete transport implementations.
	//
	// 返回一个 chan ，每当本地 tcp 端口收到一个连接，通过该 chan 交由上层处理.
	StreamCh() <-chan net.Conn


	// Shutdown is called when memberlist is shutting down; this gives the
	// transport a chance to clean up any listeners.
	//
	// 退出时，调用该函数，释放相关通信资源。
	Shutdown() error


}
