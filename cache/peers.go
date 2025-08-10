package cache

import "github.com/LudensCS/Cache/cache/cachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(Req *cachepb.Request, Resp *cachepb.Response) error
}
