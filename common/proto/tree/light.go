package tree

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/utils/uuid"
)

// N is the extracted interface from Node
type N interface {
	GetUuid() string
	GetPath() string
	GetType() NodeType
	GetSize() int64
	GetMTime() int64
	GetMode() int32
	GetEtag() string
	IsLeaf() bool

	UpdatePath(p string)
	UpdateUuid(u string)
	UpdateEtag(e string)
	UpdateSize(s int64)
	UpdateMTime(s int64)
	UpdateMode(s int32)
	SetType(NodeType)
	RenewUuidIfEmpty(force bool)

	SetChildrenSize(uint64)
	SetChildrenFiles(uint64)
	SetChildrenFolders(uint64)
	GetChildrenSize() (uint64, bool)
	GetChildrenFiles() (uint64, bool)
	GetChildrenFolders() (uint64, bool)
	SetRawMetadata(map[string]string)
	ListRawMetadata() map[string]string

	Zap(key ...string) zapcore.Field
	ZapPath() zapcore.Field
	ZapUuid() zapcore.Field

	AsProto() *Node
}

// lightNode is a memory optimized-struct implementing the N interface.
// Beware of not reordering fields as they are memory-aligned
//
//	type lightNode struct {
//	 uuid               string             ■ ■ ■ ■ ■ ■ ■ ■
//	                                       ■ ■ ■ ■ ■ ■ ■ ■
//	 path               string             ■ ■ ■ ■ ■ ■ ■ ■
//	                                       ■ ■ ■ ■ ■ ■ ■ ■
//	 size               uint64             ■ ■ ■ ■ ■ ■ ■ ■
//	 mtime              uint64             ■ ■ ■ ■ ■ ■ ■ ■
//	 etag               string             ■ ■ ■ ■ ■ ■ ■ ■
//	                                       ■ ■ ■ ■ ■ ■ ■ ■
//	 rawMeta            map[string]string  ■ ■ ■ ■ ■ ■ ■ ■
//	 childrenSize       uint64             ■ ■ ■ ■ ■ ■ ■ ■
//	 childrenFiles      uint64             ■ ■ ■ ■ ■ ■ ■ ■
//	 childrenFolders    uint64             ■ ■ ■ ■ ■ ■ ■ ■
//	 mode               uint32             ■ ■ ■ ■
//	 nodeType           tree.NodeType              ■ ■ ■ ■
//	 leaf               bool               ■
//	 childrenSizeSet    bool                 ■
//	 childrenFilesSet   bool                   ■
//	 childrenFoldersSet bool                     ■ □ □ □ □
//	}
type lightNode struct {
	uuid    string
	path    string
	size    uint64
	mtime   uint64
	etag    string
	rawMeta map[string]string

	childrenSize    uint64
	childrenFiles   uint64
	childrenFolders uint64

	mode               uint32
	nodeType           NodeType
	childrenSizeSet    bool
	childrenFilesSet   bool
	childrenFoldersSet bool
}

// LightNode create a lightNode
func LightNode(nodeType NodeType, uuid, path, eTag string, size, mTime int64, mode int32) N {
	return &lightNode{
		nodeType: nodeType,
		uuid:     uuid,
		path:     path,
		etag:     eTag,
		size:     uint64(size),
		mtime:    uint64(mTime),
		mode:     uint32(mode),
	}
}

func LightNodeFromProto(n *Node) N {
	ln := &lightNode{
		uuid:     n.Uuid,
		path:     n.Path,
		size:     uint64(n.Size),
		mtime:    uint64(n.MTime),
		mode:     uint32(n.Mode),
		etag:     n.Etag,
		nodeType: n.Type,
	}
	if n.MetaStore != nil {
		var e error
		for k, v := range n.MetaStore {
			if k == common.MetaNamespaceNodeName || k == common.MetaNamespaceDatasourceName {
				continue
			}
			if k == common.MetaRecursiveChildrenSize {
				if ln.childrenSize, e = strconv.ParseUint(v, 10, 64); e == nil {
					ln.childrenSizeSet = true
				}
			} else if k == common.MetaRecursiveChildrenFiles {
				if ln.childrenFiles, e = strconv.ParseUint(v, 10, 64); e == nil {
					ln.childrenFilesSet = true
				}
			} else if k == common.MetaRecursiveChildrenFolders {
				if ln.childrenFolders, e = strconv.ParseUint(v, 10, 64); e == nil {
					ln.childrenFoldersSet = true
				}
			} else {
				if ln.rawMeta == nil {
					ln.rawMeta = make(map[string]string)
				}
				ln.rawMeta[k] = v
			}
		}
	}
	/*
		if len(ln.rawMeta) > 0 {
			fmt.Println("Raw meta not empty", ln.rawMeta)
		}
	*/
	return ln
}

func (l *lightNode) GetUuid() string {
	return l.uuid
}

func (l *lightNode) GetPath() string {
	return l.path
}

func (l *lightNode) GetType() NodeType {
	return l.nodeType
}

func (l *lightNode) GetSize() int64 {
	return int64(l.size)
}

func (l *lightNode) GetMTime() int64 {
	return int64(l.mtime)
}

func (l *lightNode) GetMode() int32 {
	return int32(l.mode)
}

func (l *lightNode) GetEtag() string {
	return l.etag
}

func (l *lightNode) IsLeaf() bool {
	return l.nodeType == NodeType_LEAF
}

func (l *lightNode) UpdatePath(p string) {
	l.path = p
}

func (l *lightNode) UpdateUuid(u string) {
	l.uuid = u
}

func (l *lightNode) UpdateEtag(e string) {
	l.etag = e
}

func (l *lightNode) UpdateSize(s int64) {
	l.size = uint64(s)
}

func (l *lightNode) UpdateMTime(s int64) {
	l.mtime = uint64(s)
}

func (l *lightNode) UpdateMode(s int32) {
	l.mode = uint32(s)
}

func (l *lightNode) SetType(t NodeType) {
	l.nodeType = t
}

func (l *lightNode) RenewUuidIfEmpty(force bool) {
	if l.uuid == "" || force {
		l.uuid = uuid.New()
	}
}

func (l *lightNode) SetChildrenSize(u uint64) {
	l.childrenSize = u
	l.childrenSizeSet = true
}

func (l *lightNode) SetChildrenFiles(u uint64) {
	l.childrenFiles = u
	l.childrenFilesSet = true
}

func (l *lightNode) SetChildrenFolders(u uint64) {
	l.childrenFolders = u
	l.childrenFoldersSet = true
}

func (l *lightNode) GetChildrenSize() (uint64, bool) {
	return l.childrenSize, l.childrenSizeSet
}

func (l *lightNode) GetChildrenFiles() (uint64, bool) {
	return l.childrenFiles, l.childrenFilesSet
}

func (l *lightNode) GetChildrenFolders() (uint64, bool) {
	return l.childrenFolders, l.childrenFoldersSet
}

func (l *lightNode) SetRawMetadata(m map[string]string) {
	if l.rawMeta == nil {
		l.rawMeta = make(map[string]string, len(m))
	}
	for k, v := range m {
		l.rawMeta[k] = v
	}
}

func (l *lightNode) ListRawMetadata() map[string]string {
	return l.rawMeta
}

func (l *lightNode) Zap(key ...string) zapcore.Field {
	if len(key) > 0 {
		return zap.Object(key[0], l)
	} else {
		return zap.Object(common.KeyNode, l)
	}
}

func (l *lightNode) ZapPath() zapcore.Field {
	return zap.String("path", l.path)
}

func (l *lightNode) ZapUuid() zapcore.Field {
	return zap.String("uuid", l.uuid)
}

func (l *lightNode) AsProto() *Node {
	tn := &Node{
		Uuid:      l.uuid,
		Path:      l.path,
		Type:      l.nodeType,
		Size:      int64(l.size),
		MTime:     int64(l.mtime),
		Mode:      int32(l.mode),
		Etag:      l.etag,
		MetaStore: l.rawMeta,
	}
	if l.childrenFilesSet || l.childrenFoldersSet || l.childrenSizeSet {
		if tn.MetaStore == nil {
			tn.MetaStore = make(map[string]string)
		}
		if l.childrenSizeSet {
			tn.MetaStore[common.MetaRecursiveChildrenSize] = fmt.Sprintf("%d", l.childrenSize)
		}
		if l.childrenFilesSet {
			tn.MetaStore[common.MetaRecursiveChildrenFiles] = fmt.Sprintf("%d", l.childrenFiles)
		}
		if l.childrenFoldersSet {
			tn.MetaStore[common.MetaRecursiveChildrenFolders] = fmt.Sprintf("%d", l.childrenFolders)
		}
	}
	return tn
}

// MarshalLogObject implements custom marshalling for logs
func (l *lightNode) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	if l == nil {
		return nil
	}
	if l.uuid != "" {
		encoder.AddString("Uuid", l.uuid)
	}
	if l.path != "" {
		encoder.AddString("Path", l.path)
	}
	if l.etag != "" {
		encoder.AddString("Etag", l.etag)
	}
	if l.mtime > 0 {
		encoder.AddTime("MTime", time.Unix(int64(l.mtime), 0))
	}
	if l.size > 0 {
		encoder.AddUint64("Size", l.size)
	}
	if l.rawMeta != nil {
		_ = encoder.AddReflected("MetaStore", l.rawMeta)
	}
	return nil
}
