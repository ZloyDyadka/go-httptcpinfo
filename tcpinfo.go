package tcpinfo

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// Got from https://github.com/m-lab/tcp-info
// TCPInfo is the linux defined structure returned in RouteAttr DIAG_INFO messages.
// It corresponds to the struct tcp_info in include/uapi/linux/tcp.h
type TCPInfo struct {
	State       uint8 `json:"state"`
	CAState     uint8 `json:"ca_state"`
	Retransmits uint8 `json:"retransmits"`
	Probes      uint8 `json:"probes"`
	Backoff     uint8 `json:"backoff"`
	Options     uint8 `json:"options"`
	WScale      uint8 `json:"w_scale"`     //bit fields snd_wscale : 4, tcpi_rcv_wscale : 4;
	AppLimited  uint8 `json:"app_limited"` //delivery_rate_app_limited:1;

	RTO    uint32 `json:"rto"` // offset 8
	ATO    uint32 `json:"ato"`
	SndMSS uint32 `json:"snd_mss"`
	RcvMSS uint32 `json:"rcv_mss"`

	Unacked uint32 `json:"unacked"` // offset 24
	Sacked  uint32 `json:"sacked"`
	Lost    uint32 `json:"lost"`
	Retrans uint32 `json:"retrans"`
	Fackets uint32 `json:"f_ackets"`

	/* Times. */
	// These seem to be elapsed time, so they increase on almost every sample.
	// We can probably use them to get more info about intervals between samples.
	LastDataSent uint32 `json:"last_data_sent"` // offset 44
	LastAckSent  uint32 `json:"last_ack_sent"`  /* Not remembered, sorry. */ // offset 48
	LastDataRecv uint32 `json:"last_data_recv"` // offset 52
	LastAckRecv  uint32 `json:"last_ack_recv"`  // offset 56

	/* Metrics. */
	PMTU        uint32 `json:"p_mtu"`
	RcvSsThresh uint32 `json:"rcv_ss_thresh"`
	RTT         uint32 `json:"rtt"`
	RTTVar      uint32 `json:"rtt_var"`
	SndSsThresh uint32 `json:"snd_ss_thresh"`
	SndCwnd     uint32 `json:"snd_cwnd"`
	AdvMSS      uint32 `json:"adv_mss"`
	Reordering  uint32 `json:"reordering"`

	RcvRTT   uint32 `json:"rcv_rtt"`
	RcvSpace uint32 `json:"rcv_space"`

	TotalRetrans uint32 `json:"total_retrans"`

	PacingRate    int64 `json:"pacing_rate"`     // This is often -1, so better for it to be signed
	MaxPacingRate int64 `json:"max_pacing_rate"` // This is often -1, so better to be signed.

	// NOTE: In linux, these are uint64, but we make them int64 here for compatibility with BigQuery
	BytesAcked    int64 `json:"bytes_acked"`    /* RFC4898 tcpEStatsAppHCThruOctetsAcked */
	BytesReceived int64 `json:"bytes_received"` /* RFC4898 tcpEStatsAppHCThruOctetsReceived */
	SegsOut       int32 `json:"segs_out"`       /* RFC4898 tcpEStatsPerfSegsOut */
	SegsIn        int32 `json:"segs_in"`        /* RFC4898 tcpEStatsPerfSegsIn */

	NotsentBytes uint32 `json:"notsent_bytes"`
	MinRTT       uint32 `json:"min_rtt"`
	DataSegsIn   uint32 `json:"data_segs_in"`  /* RFC4898 tcpEStatsDataSegsIn */
	DataSegsOut  uint32 `json:"data_segs_out"` /* RFC4898 tcpEStatsDataSegsOut */

	// NOTE: In linux, this is uint64, but we make it int64 here for compatibility with BigQuery
	DeliveryRate int64 `json:"delivery_rate"`

	BusyTime      int64 `json:"busy_time"`       /* Time (usec) busy sending data */
	RWndLimited   int64 `json:"r_wnd_limited"`   /* Time (usec) limited by receive window */
	SndBufLimited int64 `json:"snd_buf_limited"` /* Time (usec) limited by send buffer */

	Delivered   uint32 `json:"delivered"`
	DeliveredCE uint32 `json:"delivered_ce"`

	// NOTE: In linux, these are uint64, but we make them int64 here for compatibility with BigQuery
	BytesSent    int64 `json:"bytes_sent"`    /* RFC4898 tcpEStatsPerfHCDataOctetsOut */
	BytesRetrans int64 `json:"bytes_retrans"` /* RFC4898 tcpEStatsPerfOctetsRetrans */

	DSackDups uint32 `json:"d_sack_dups"` /* RFC4898 tcpEStatsStackDSACKDups */
	ReordSeen uint32 `json:"reord_seen"`  /* reordering events seen */
}

type WScale struct {
	Send uint8
	Recv uint8
}

// UnpackWScale unpacks WScale fields from TCPInfo.WScale
func UnpackWScale(ws uint8) WScale {
	return WScale{
		Send: ws & 0x0F,
		Recv: ws >> 4,
	}
}

// GetTCPInfoByFD returns TCPInfo by file descriptor
func GetTCPInfoByFD(fd uintptr) (*TCPInfo, error) {
	info, err := getsockoptTCPInfo(int(fd), syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		return nil, errors.Wrap(err, "get tcp info socket options")
	}

	return info, nil
}

func getsockoptTCPInfo(fd, level, opt int) (*TCPInfo, error) {
	var value TCPInfo
	vallen := uint32(unsafe.Sizeof(value))
	err := getSockOpt(fd, level, opt, unsafe.Pointer(&value), &vallen)
	return &value, err
}

func getSockOpt(s int, level int, name int, val unsafe.Pointer, vallen *uint32) error {
	_, _, e1 := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(vallen)), 0)
	if e1 != 0 {
		return e1
	}

	return nil
}
