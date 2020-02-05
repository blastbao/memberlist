package memberlist

import (
	"io"
	"log"
	"os"
	"time"
)

type Config struct {



	// The name of this node. This must be unique in the cluster.
	// 节点名称，是节点在集群内的唯一标识。
	Name string


	// Transport is a hook for providing custom code to communicate with other nodes.
	// If this is left nil, then memberlist will by default make a NetTransport
	// using BindAddr and BindPort from this structure.
	//
	// Transport 是同集群内的节点通信的抽象，包括tcp、udp。
	// 不配置这个接口，默认使用 memberlist 提供的 NetTransport 。
	// 这个基本不用配置，只有官方提供的功能满足不了你的时候，把源码理解透了，再考虑自定义这个接口。
	Transport Transport





	// Configuration related to what address to bind to and ports to listen on.
	// The port is used for both UDP and TCP gossip.
	//
	// It is assumed other nodes are running on this port, but they do not need to.
	//
	// 本地绑定的 IP 和监听的 Port，Port 同时支持 udp 和 tcp 。
	//
	// 在 gossip 集群中，新节点可以通过集群中任意一个节点的 BindAddr、BindPort 加入到集群中。
	// 默认为 "0.0.0.0:7946"，BindPort 如果配置成 0 ，memberlist 会动态绑定一个端口。
	BindAddr string
	BindPort int





	// Configuration related to what address to advertise to other cluster members.
	//
	// Used for nat traversal.
	//
	//
	// 这对 ip、port 是集群其它节点与自己通讯用的，也就是外部可以访问到的 ip 和 port。
	// 上边的 BindAddr、BindPort 是本地监听的 ip 和 port（eg. "0.0.0.0:7946"），
	// 但实际本机的出口地址是 "192.168.1.100:7946"，那么 AdvertiseAddr 应该是 192.168.1.100，AdvertisePort 是 7946，
	// 否则其他节点不知道你的真实 ip。
	//
	// 可能带来误解，这个又要根据 AdvertiseAddr、AdvertisePort 在本地启一个什么服务，其实不是，默认是空，
	// memberlist 会解析到节点绑定的 ip 和 port， nat 转换后的 ip 和 port 。
	AdvertiseAddr string
	AdvertisePort int



	// ProtocolVersion is the configured protocol version that we will _speak_.
	// This must be between ProtocolVersionMin and ProtocolVersionMax.
	//
	// 协议版本。现在有五个版本，差别不大，例如 v3 加了 tcp ping，v4 支持间接 ping，默认使用 v2 版本。
	ProtocolVersion uint8



	// TCPTimeout is the timeout for establishing a stream connection with
	// a remote node for a full state sync, and for stream read and write
	// operations. This is a legacy name for backwards compatibility, but
	// should really be called StreamTimeout now that we have generalized
	// the transport.
	//
	// 建立 tcp 链接的超时时间，根据网络情况配置即可。
	TCPTimeout time.Duration




	// IndirectChecks is the number of nodes that will be asked to perform
	// an indirect probe of a node in the case a direct probe fails. Memberlist
	// waits for an ack from any single indirect node, so increasing this
	// number will increase the likelihood that an indirect probe will succeed
	// at the expense of bandwidth.
	//
	// v4 协议支持间接探测。当节点启动之后，每隔一定的时间间隔，会选取一个节点对其发送一个 Ping(UDP) 消息，
	// 当 Ping 消息失败后，会随机选取 IndirectChecks 个节点发起间接的 Ping。
	IndirectChecks int




	// RetransmitMult is the multiplier for the number of retransmissions
	// that are attempted for messages broadcasted over gossip. The actual
	// count of retransmissions is calculated using the formula:
	//
	//   Retransmits = RetransmitMult * log(N+1)
	//
	// This allows the retransmits to scale properly with cluster size. The
	// higher the multiplier, the more likely a failed broadcast is to converge
	// at the expense of increased bandwidth.
	//
	//
	// 广播队列里的消息发送失败超过一定次数后，消息就会被丢弃，RetransmitMult 就是用来计算重传次数的，
	// Retransmits = RetransmitMult * log(N+1)
	RetransmitMult int






	// 探测某个节点超时后，会将该节点标记为 suspect 节点，节点被标记为 suspect 后，本地启动一个定时器，
	// 而后发出一个 suspect 广播，在超时时间内，如果收到其他节点发送过来的 suspect 消息，就将本地的 suspect 确认数加1，
	// 当定时器到期后，suspect 确认数达到要求并且该节依旧不是 alive 状态，会将该节点标记 dead。
	//
	// 这个时间计算方式如下：
	// 	1. SuspicionTimeout = SuspicionMult * log(N+1) * ProbeInterval
	// 	2. SuspicionMaxTimeout = SuspicionMaxTimeoutMult * SuspicionTimeout
	//
	// 这两个值一般也不要我们配置，使用默认即可


	// SuspicionMult is the multiplier for determining the time an
	// inaccessible node is considered suspect before declaring it dead.
	// The actual timeout is calculated using the formula:
	//
	//   SuspicionTimeout = SuspicionMult * log(N+1) * ProbeInterval
	//
	// This allows the timeout to scale properly with expected propagation
	// delay with a larger cluster size. The higher the multiplier, the longer
	// an inaccessible node is considered part of the cluster before declaring
	// it dead, giving that suspect node more time to refute if it is indeed
	// still alive.
	SuspicionMult int





	// SuspicionMaxTimeoutMult is the multiplier applied to the
	// SuspicionTimeout used as an upper bound on detection time. This max
	// timeout is calculated using the formula:
	//
	// SuspicionMaxTimeout = SuspicionMaxTimeoutMult * SuspicionTimeout
	//
	// If everything is working properly, confirmations from other nodes will
	// accelerate suspicion timers in a manner which will cause the timeout
	// to reach the base SuspicionTimeout before that elapses, so this value
	// will typically only come into play if a node is experiencing issues
	// communicating with other nodes. It should be set to a something fairly
	// large so that a node having problems will have a lot of chances to
	// recover before falsely declaring other nodes as failed, but short
	// enough for a legitimately isolated node to still make progress marking
	// nodes failed in a reasonable amount of time.
	SuspicionMaxTimeoutMult int




	// PushPullInterval is the interval between complete state syncs.
	// Complete state syncs are done with a single node over TCP and are
	// quite expensive relative to standard gossiped messages. Setting this
	// to zero will disable state push/pull syncs completely.
	//
	// Setting this interval lower (more frequent) will increase convergence
	// speeds across larger clusters at the expense of increased bandwidth
	// usage.
	//
	//
	// 每隔 PushPullInterval 时间间隔，随机选取一个节点，跟它建立 tcp 连接，
	// 然后将本节点的全部状态通过 tcp 传到对方，对方也把它的状态传送回来，进行状态同步。
	// 此值根据网络情况进行调整即可。
	PushPullInterval time.Duration




	// 探测间隔 和 探测超时时间。根据网络情况进行调整即可。


	// ProbeInterval and ProbeTimeout are used to configure probing
	// behavior for memberlist.
	//
	// ProbeInterval is the interval between random node probes. Setting
	// this lower (more frequent) will cause the memberlist cluster to detect
	// failed nodes more quickly at the expense of increased bandwidth usage.
	//
	// ProbeTimeout is the timeout to wait for an ack from a probed node
	// before assuming it is unhealthy. This should be set to 99-percentile
	// of RTT (round-trip time) on your network.
	ProbeInterval time.Duration
	ProbeTimeout  time.Duration

	// DisableTcpPings will turn off the fallback TCP pings that are attempted
	// if the direct UDP ping fails. These get pipelined along with the
	// indirect UDP pings.
	//
	// 关闭 tcp ping 。这个参数一般也不用管它。
	DisableTcpPings bool

	// AwarenessMaxMultiplier will increase the probe interval if the node
	// becomes aware that it might be degraded and not meeting the soft real
	// time requirements to reliably probe other nodes.
	//
	// 在节点认为自己不能可靠的探测其他节点时，会根据这个参数增加探测间隔。一般使用默认配置即可。
	AwarenessMaxMultiplier int

	// GossipInterval and GossipNodes are used to configure the gossip
	// behavior of memberlist.
	//
	// GossipInterval is the interval between sending messages that need
	// to be gossiped that haven't been able to piggyback on probing messages.
	// If this is set to zero, non-piggyback gossip is disabled. By lowering
	// this value (more frequent) gossip messages are propagated across
	// the cluster more quickly at the expense of increased bandwidth.
	//
	// GossipNodes is the number of random nodes to send gossip messages to
	// per GossipInterval. Increasing this number causes the gossip messages
	// to propagate across the cluster more quickly at the expense of
	// increased bandwidth.
	//
	// GossipToTheDeadTime is the interval after which a node has died that
	// we will still try to gossip to it. This gives it a chance to refute.
	//
	//
	GossipInterval      time.Duration	// 检查广播队列是否有数据需要发送给其他节点的时间间隔
	GossipNodes         int				// 每次给几个节点扩散数据
	GossipToTheDeadTime time.Duration	// 在这个时间内仍然会尝试给 Dead 状态的节点发送数据，使用默认配置即可




	// GossipVerifyIncoming、GossipVerifyOutgoing 标识了是否对网络上的数据进行加密。
	// 当取值为 true 时，如果加密失败则报错；为 false 时，如果加密失败则明文传输。


	// GossipVerifyIncoming controls whether to enforce encryption for incoming
	// gossip. It is used for upshifting from unencrypted to encrypted gossip on
	// a running cluster.
	GossipVerifyIncoming bool

	// GossipVerifyOutgoing controls whether to enforce encryption for outgoing
	// gossip. It is used for upshifting from unencrypted to encrypted gossip on
	// a running cluster.
	GossipVerifyOutgoing bool




	// EnableCompression is used to control message compression. This can
	// be used to reduce bandwidth usage at the cost of slightly more CPU
	// utilization. This is only available starting at protocol version 1.
	//
	//
	// 是否数据压缩，默认是 true ，可以减少带宽。根据需要来配置。
	EnableCompression bool



	// SecretKey is used to initialize the primary encryption key in a keyring.
	// The primary encryption key is the only key used to encrypt messages and
	// the first key used while attempting to decrypt messages. Providing a
	// value for this primary key will enable message-level encryption and
	// verification, and automatically install the key onto the keyring.
	// The value should be either 16, 24, or 32 bytes to select AES-128,
	// AES-192, or AES-256.
	SecretKey []byte

	// The keyring holds all of the encryption keys used internally. It is
	// automatically initialized using the SecretKey and SecretKeys values.
	Keyring *Keyring





	// Delegate and Events are delegates for receiving and providing
	// data to memberlist via callback mechanisms. For Delegate, see
	// the Delegate interface. For Events, see the EventDelegate interface.
	//
	// The DelegateProtocolMin/Max are used to guarantee protocol-compatibility
	// for any custom messages that the delegate might do (broadcasts,
	// local/remote state, etc.). If you don't set these, then the protocol
	// versions will just be zero, and version compliance won't be done.
	Delegate                Delegate
	DelegateProtocolVersion uint8
	DelegateProtocolMin     uint8
	DelegateProtocolMax     uint8


	// 当集群内有节点加入、离开、元数据变更都会触发 EventDelegate 事件回调
	Events                  EventDelegate

	// 当 `Name` 和 `AdvertiseAddr`/`AdvertisePort` 不对应时，会触发 ConflictDelegate 回调
	Conflict                ConflictDelegate

	// 当某节点 Join 该集群时，集群内的每个节点会触发 MergeDelegate 的接口一次，用于接受新节点的数据。该接口可以是nil。
	Merge                   MergeDelegate

	// 当探测其他节点时，会触发相关 ping 接口
	Ping                    PingDelegate

	//
	Alive                   AliveDelegate

	// DNSConfigPath points to the system's DNS config file, usually located
	// at /etc/resolv.conf. It can be overridden via config for easier testing.
	//
	// 指向系统的 DNS 配置文件，linux就是 "/etc/resolv.conf"。
	DNSConfigPath string



	// 定义 memerlist 的日志输出方式


	// LogOutput is the writer where logs should be sent. If this is not
	// set, logging will go to stderr by default. You cannot specify both LogOutput
	// and Logger at the same time.
	LogOutput io.Writer

	// Logger is a custom logger which you provide. If Logger is set, it will use
	// this for the internal logger. If Logger is not set, it will fall back to the
	// behavior for using LogOutput. You cannot specify both LogOutput and Logger
	// at the same time.
	Logger *log.Logger

	// Size of Memberlist's internal channel which handles UDP messages. The
	// size of this determines the size of the queue which Memberlist will keep
	// while UDP messages are handled.
	//
	// 广播队列的大小，默认1024，基本足够，不需要改动。
	HandoffQueueDepth int


	// Maximum number of bytes that memberlist will put in a packet (this
	// will be for UDP packets by default with a NetTransport). A safe value
	// for this is typically 1400 bytes (which is the default). However,
	// depending on your network's MTU (Maximum Transmission Unit) you may
	// be able to increase this to get more content into each gossip packet.
	// This is a legacy name for backward compatibility but should really be
	// called PacketBufferSize now that we have generalized the transport.
	UDPBufferSize int


	// DeadNodeReclaimTime controls the time before a dead node's name can be
	// reclaimed by one with a different address or port. By default, this is 0,
	// meaning nodes cannot be reclaimed this way.
	//
	// 如果配置了，假如节点处于 dead 状态，DeadNodeReclaimTime 时间范围内允许节点 name 相同，
	// 但 ip、port 不同的节点对其对其进行状态更新。
	//
	// 默认是 0，此时如果发生上述现象则发生冲突，不理解没关系，不配置就可以了。
	DeadNodeReclaimTime time.Duration

}

// DefaultLANConfig returns a sane set of configurations for Memberlist.
// It uses the hostname as the node name, and otherwise sets very conservative
// values that are sane for most LAN environments. The default configuration
// errs on the side of caution, choosing values that are optimized
// for higher convergence at the cost of higher bandwidth usage. Regardless,
// these values are a good starting point when getting started with memberlist.
func DefaultLANConfig() *Config {
	hostname, _ := os.Hostname()
	return &Config{
		Name:                    hostname,
		BindAddr:                "0.0.0.0",
		BindPort:                7946,
		AdvertiseAddr:           "",
		AdvertisePort:           7946,
		ProtocolVersion:         ProtocolVersion2Compatible,
		TCPTimeout:              10 * time.Second,       // Timeout after 10 seconds
		IndirectChecks:          3,                      // Use 3 nodes for the indirect ping
		RetransmitMult:          4,                      // Retransmit a message 4 * log(N+1) nodes
		SuspicionMult:           4,                      // Suspect a node for 4 * log(N+1) * Interval
		SuspicionMaxTimeoutMult: 6,                      // For 10k nodes this will give a max timeout of 120 seconds
		PushPullInterval:        30 * time.Second,       // Low frequency
		ProbeTimeout:            500 * time.Millisecond, // Reasonable RTT time for LAN
		ProbeInterval:           1 * time.Second,        // Failure check every second
		DisableTcpPings:         false,                  // TCP pings are safe, even with mixed versions
		AwarenessMaxMultiplier:  8,                      // Probe interval backs off to 8 seconds

		GossipNodes:          3,                      // Gossip to 3 nodes
		GossipInterval:       200 * time.Millisecond, // Gossip more rapidly
		GossipToTheDeadTime:  30 * time.Second,       // Same as push/pull
		GossipVerifyIncoming: true,
		GossipVerifyOutgoing: true,

		EnableCompression: true, // Enable compression by default

		SecretKey: nil,
		Keyring:   nil,

		DNSConfigPath: "/etc/resolv.conf",

		HandoffQueueDepth: 1024,
		UDPBufferSize:     1400,
	}
}

// DefaultWANConfig works like DefaultConfig, however it returns a configuration
// that is optimized for most WAN environments.
//
// The default configuration is still very conservative and errs on the side of caution.
func DefaultWANConfig() *Config {

	conf := DefaultLANConfig()
	conf.TCPTimeout = 30 * time.Second
	conf.SuspicionMult = 6
	conf.PushPullInterval = 60 * time.Second
	conf.ProbeTimeout = 3 * time.Second
	conf.ProbeInterval = 5 * time.Second
	conf.GossipNodes = 4 // Gossip less frequently, but to an additional node
	conf.GossipInterval = 500 * time.Millisecond
	conf.GossipToTheDeadTime = 60 * time.Second

	return conf
}

// DefaultLocalConfig works like DefaultConfig, however it returns a configuration
// that is optimized for a local loopback environments.
//
// The default configuration is still very conservative and errs on the side of caution.
func DefaultLocalConfig() *Config {

	conf := DefaultLANConfig()
	conf.TCPTimeout = time.Second
	conf.IndirectChecks = 1
	conf.RetransmitMult = 2
	conf.SuspicionMult = 3
	conf.PushPullInterval = 15 * time.Second
	conf.ProbeTimeout = 200 * time.Millisecond
	conf.ProbeInterval = time.Second
	conf.GossipInterval = 100 * time.Millisecond
	conf.GossipToTheDeadTime = 15 * time.Second
	return conf
}

// Returns whether or not encryption is enabled
func (c *Config) EncryptionEnabled() bool {
	return c.Keyring != nil && len(c.Keyring.GetKeys()) > 0
}
